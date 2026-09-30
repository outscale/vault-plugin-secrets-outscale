package outscalevault

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/iso8601"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
)

func pathRootCreds(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "root-creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
			"ttl": {
				Type:        framework.TypeDurationSecond,
				Description: "TTL for the generated secret. If not set, uses the role's TTL.",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathRootCredsRead,
			logical.UpdateOperation: b.pathRootCredsRead,
		},
		HelpSynopsis:    pathRootCredsHelpSynopsis,
		HelpDescription: pathRootCredsHelpDescription,
	}
}

func (b *backend) pathRootCredsRead(
	ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	roleEntry, err := b.getRootRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, fmt.Errorf("error retriving role: %w", err)
	}

	if roleEntry == nil {
		return logical.ErrorResponse("root role %q does not exist", roleName), nil
	}

	var ttl time.Duration
	if ttlRaw, ok := d.GetOk("ttl"); ok {
		ttl = time.Duration(ttlRaw.(int)) * time.Second
	}

	return b.createRootCreds(ctx, req, roleEntry, roleName, ttl)
}

func (b *backend) createRootCreds(
	ctx context.Context,
	req *logical.Request,
	roleEntry *outscaleRootRoleEntry,
	roleName string,
	ttlOverride time.Duration,
) (*logical.Response, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return b.responseFromErr("creating root credentials", err)
	}

	ttl := roleEntry.TTL
	if ttlOverride > 0 {
		ttl = ttlOverride
	}

	tag := fmt.Sprintf("vault:root-creds/%s", roleName)
	// Expiration date is set to TTL + 1 minute to let Vault revoke the credential before expiration
	createAccessKey, err := client.CreateAccessKey(ctx, osc.CreateAccessKeyRequest{
		ExpirationDate: &iso8601.Time{Time: time.Now().Add(ttl).Add(time.Minute)},
		Tag:            &tag,
	})
	if err != nil {
		return b.responseFromErr("creating root credentials", err)
	}

	resp := b.Secret(secretRootAccessKeyType).Response(map[string]any{
		"access_key": *createAccessKey.AccessKey.AccessKeyId,
		"secret_key": *createAccessKey.AccessKey.SecretKey,
	}, map[string]any{
		"access_key": *createAccessKey.AccessKey.AccessKeyId,
		"role":       roleName,
	})

	resp.Secret.TTL = ttl
	resp.Secret.Renewable = false

	return resp, nil
}

func secretRootAccessKeys(b *backend) *framework.Secret {
	return &framework.Secret{
		Type: secretRootAccessKeyType,
		Fields: map[string]*framework.FieldSchema{
			"access_key": {
				Type:        framework.TypeString,
				Description: "Access Key",
			},
			"secret_key": {
				Type:        framework.TypeString,
				Description: "Secret Key",
			},
		},
		Revoke: b.secretRootAccessKeyRevoke,
	}
}

func (b *backend) secretRootAccessKeyRevoke(
	ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	// Get the access key from the internal data
	akRaw, ok := req.Secret.InternalData["access_key"]
	if !ok {
		return nil, fmt.Errorf("secret is missing username internal data")
	}
	accessKey, ok := akRaw.(string)
	if !ok {
		return nil, fmt.Errorf("secret is missing username internal data")
	}

	b.lock.RLock()
	defer b.lock.RUnlock()

	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return b.responseFromErr("revoking root credentials", err)
	}

	_, err = client.UpdateAccessKey(ctx, osc.UpdateAccessKeyRequest{
		AccessKeyId: accessKey,
		State:       "INACTIVE",
	})
	if err != nil {
		switch {
		case osc.HasErrorCode(err, []string{"5076", "5099", "4000"}):
		// AK expired or already deleted
		default:
			return b.responseFromErr("deactivating root credentials", err)
		}
	}

	_, err = client.DeleteAccessKey(ctx, osc.DeleteAccessKeyRequest{
		AccessKeyId: accessKey,
	})
	if err != nil {
		switch {
		case osc.HasErrorCode(err, []string{"5076", "5099", "4000"}):
		// AK expired or already deleted
		default:
			return b.responseFromErr("deleting root credentials", err)
		}
	}

	return nil, nil
}

const (
	pathRootCredsHelpDescription = ""
	pathRootCredsHelpSynopsis    = ""
	secretRootAccessKeyType      = "root_access_keys"
)
