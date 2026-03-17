package outscalevault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type outscaleRootRoleEntry struct {
	TTL time.Duration `json:"ttl"`
}

func (r *outscaleRootRoleEntry) toResponseData() map[string]any {
	respData := map[string]any{
		"ttl": r.TTL.Seconds(),
	}
	return respData
}

func pathRootRole(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "root-role/" + framework.GenericNameRegex("name"),
			Fields: map[string]*framework.FieldSchema{
				"name": {
					Type:        framework.TypeLowerCaseString,
					Description: "Name of the role",
					Required:    true,
				},
				"ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Default lease for generated credentials. If not set or set to 0, will use system default.",
					Default:     2 * time.Hour,
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathRootRolesRead,
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathRootRolesWrite,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRootRolesWrite,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathRootRolesDelete,
				},
			},
			ExistenceCheck:  b.pathRootRoleExistenceCheck,
			HelpSynopsis:    pathRootRoleHelpSynopsis,
			HelpDescription: pathRootRoleHelpDescription,
		},
		{
			Pattern: "root-role/?$",
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathRootRolesList,
				},
			},
			HelpSynopsis:    pathRootRoleListHelpSynopsis,
			HelpDescription: pathRootRoleListHelpDescription,
		},
	}
}

func (b *backend) pathRootRolesList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, "root-role/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

func (b *backend) pathRootRolesRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entry, err := b.getRootRole(ctx, req.Storage, d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: entry.toResponseData(),
	}, nil
}

func (b *backend) pathRootRolesWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name, ok := d.GetOk("name")
	if !ok {
		return logical.ErrorResponse("missing name"), nil
	}

	roleEntry, err := b.getRootRole(ctx, req.Storage, name.(string))
	if err != nil {
		return nil, err
	}

	if roleEntry == nil {
		roleEntry = &outscaleRootRoleEntry{}
	}

	createOperation := (req.Operation == logical.CreateOperation)

	if ttlRaw, ok := d.GetOk("ttl"); ok {
		roleEntry.TTL = time.Duration(ttlRaw.(int)) * time.Second
	} else if createOperation {
		roleEntry.TTL = time.Duration(d.Get("ttl").(int)) * time.Second
	}

	if err := setRootRole(ctx, req.Storage, name.(string), roleEntry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathRootRolesDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "root-role/"+d.Get("name").(string))
	if err != nil {
		return nil, fmt.Errorf("error deleting root-role: %w", err)
	}

	return nil, nil
}

func setRootRole(ctx context.Context, s logical.Storage, name string, roleEntry *outscaleRootRoleEntry) error {
	entry, err := logical.StorageEntryJSON("root-role/"+name, roleEntry)
	if err != nil {
		return err
	}

	if entry == nil {
		return fmt.Errorf("failed to create storage entry for root role")
	}

	if err := s.Put(ctx, entry); err != nil {
		return err
	}

	return nil
}

func (b *backend) getRootRole(ctx context.Context, s logical.Storage, name string) (*outscaleRootRoleEntry, error) {
	if name == "" {
		return nil, fmt.Errorf("missing root-role name")
	}

	entry, err := s.Get(ctx, "root-role/"+name)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var role outscaleRootRoleEntry

	if err := entry.DecodeJSON(&role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (b *backend) pathRootRoleExistenceCheck(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (bool, error) {
	out, err := b.getRootRole(ctx, req.Storage, data.Get("name").(string))
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

const (
	pathRootRoleHelpSynopsis    = ``
	pathRootRoleHelpDescription = ``

	pathRootRoleListHelpSynopsis    = ``
	pathRootRoleListHelpDescription = ``
)
