package outscalevault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type outscaleRoleEntry struct {
	UserGroupNames    []string      `json:"user_group_names"`
	TTL               time.Duration `json:"ttl"`
	MaxTTL            time.Duration `json:"max_ttl"`
	UserNameTemplate  string        `json:"user_name_template"`
	UserEmailTemplate string        `json:"user_email_template"`
	PathTemplate      string        `json:"path_template"`
}

func (r *outscaleRoleEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"user_group_names":    r.UserGroupNames,
		"ttl":                 r.TTL.Seconds(),
		"max_ttl":             r.MaxTTL.Seconds(),
		"user_name_template":  r.UserNameTemplate,
		"user_email_template": r.UserEmailTemplate,
		"path_template":       r.PathTemplate,
	}
	return respData
}

func (b *backend) getRole(
	ctx context.Context,
	s logical.Storage,
	name string,
) (*outscaleRoleEntry, error) {
	if name == "" {
		return nil, fmt.Errorf("missing role name")
	}

	entry, err := s.Get(ctx, "roles/"+name)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var role outscaleRoleEntry

	if err := entry.DecodeJSON(&role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (b *backend) pathRolesRead(
	ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	entry, err := b.getRole(ctx, req.Storage, d.Get("name").(string))
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

func setRole(
	ctx context.Context,
	s logical.Storage,
	name string,
	roleEntry *outscaleRoleEntry,
) error {
	entry, err := logical.StorageEntryJSON("roles/"+name, roleEntry)
	if err != nil {
		return err
	}

	if entry == nil {
		return fmt.Errorf("failed to create storage entry for role")
	}

	if err := s.Put(ctx, entry); err != nil {
		return err
	}

	return nil
}

func (b *backend) pathRolesWrite(ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	name, ok := d.Get("name").(string)
	if !ok {
		return logical.ErrorResponse("missing name"), nil
	}

	roleEntry, err := b.getRole(ctx, req.Storage, name)
	if err != nil {
		return nil, err
	}

	if roleEntry == nil {
		roleEntry = &outscaleRoleEntry{}
	}

	createOperation := (req.Operation == logical.CreateOperation)

	ugm, ok := d.Get("user_group_names").([]string)
	if ok {
		roleEntry.UserGroupNames = ugm
	} else {
		return logical.ErrorResponse("user_group_names must to defined"), nil
	}

	user_name_template, ok := d.Get("user_name_template").(string)
	if ok && len(user_name_template) > 0 {
		roleEntry.UserNameTemplate = user_name_template
	} else {
		roleEntry.UserNameTemplate = "vault-user-{{random 20}}"
	}

	user_email_template, ok := d.Get("user_email_template").(string)
	if ok {
		roleEntry.UserEmailTemplate = user_email_template
	}

	path_template, ok := d.Get("path_template").(string)
	if ok {
		roleEntry.PathTemplate = path_template
	}

	if ttlRaw, ok := d.GetOk("ttl"); ok {
		roleEntry.TTL = time.Duration(ttlRaw.(int)) * time.Second
	} else if createOperation {
		roleEntry.TTL = time.Duration(d.Get("ttl").(int)) * time.Second
	}

	if maxTTLRaw, ok := d.GetOk("max_ttl"); ok {
		roleEntry.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	} else if createOperation {
		roleEntry.MaxTTL = time.Duration(d.Get("max_ttl").(int)) * time.Second
	}

	if roleEntry.MaxTTL != 0 && roleEntry.TTL > roleEntry.MaxTTL {
		return logical.ErrorResponse("ttl cannot be greater than max_ttl"), nil
	}
	if err := setRole(ctx, req.Storage, name, roleEntry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathRolesDelete(
	ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "roles/"+d.Get("name").(string))
	if err != nil {
		return nil, fmt.Errorf("error deleting hashiCups role: %w", err)
	}

	return nil, nil
}

func (b *backend) pathRolesList(
	ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, "role/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}

func (b *backend) pathRolesExistenceCheck(ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

func pathRole(b *backend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "roles/" + framework.GenericNameWithAtRegex("name"),
			Fields: map[string]*framework.FieldSchema{
				"name": {
					Type:        framework.TypeLowerCaseString,
					Description: "Name of the role",
					Required:    true,
				},
				"user_group_names": {
					Type:        framework.TypeStringSlice,
					Description: "Groups name the EIM should be associated with this role. The user groups are defined in the EIM.",
					Required:    true,
				},
				"user_name_template": {
					Type:        framework.TypeString,
					Description: "User Name template for EIM user",
				},
				"user_email_template": {
					Type:        framework.TypeString,
					Description: "User email template for EIM user",
				},
				"path_template": {
					Type:        framework.TypeString,
					Description: "path template for EIM user",
				},
				"ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Default lease for generated credentials. If not set or set to 0, will use system default.",
				},
				"max_ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Maximum time for role. If not set or set to 0, will use system default.",
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathRolesRead,
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathRolesWrite,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRolesWrite,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathRolesDelete,
				},
			},
			HelpSynopsis:    pathRoleHelpSynopsis,
			HelpDescription: pathRoleHelpDescription,
			ExistenceCheck:  b.pathRolesExistenceCheck,
		},
		{
			Pattern: "roles/?$",
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathRolesList,
				},
			},
			HelpSynopsis:    pathRoleListHelpSynopsis,
			HelpDescription: pathRoleListHelpDescription,
		},
	}
}

const (
	pathRoleHelpSynopsis        = "a"
	pathRoleHelpDescription     = "aaaa"
	pathRoleListHelpSynopsis    = "b"
	pathRoleListHelpDescription = "bbbb"
)
