package outscalevault

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
)

const (
	pathConfigRotateRootHelpDesc = "Rotates the plugin configuration access key. The previous access key cannot be restored after rotation."
	pathConfigRotateRootHelpSyn  = "Rotate the plugin configuration access key."
	pathConfigRotateRootPattern  = "config/rotate-root"
)

func pathConfigRotateRoot(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: pathConfigRotateRootPattern,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback:                    b.pathConfigRotateRootUpdate,
				ForwardPerformanceSecondary: true,
				ForwardPerformanceStandby:   true,
			},
		},
		HelpSynopsis:    pathConfigRotateRootHelpSyn,
		HelpDescription: pathConfigRotateRootHelpDesc,
	}
}

func (b *backend) pathConfigRotateRootUpdate(
	ctx context.Context,
	req *logical.Request,
	_ *framework.FieldData,
) (*logical.Response, error) {
	err := b.rotateRootCredential(ctx, req)
	if err != nil {
		return b.responseFromErr("rotating the plugin configuration access key", err)
	}

	return nil, nil
}

func (b *backend) rotateRootCredential(ctx context.Context, req *logical.Request) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.rotateRootCredentialLocked(ctx, req)
}

func (b *backend) rotateRootCredentialLocked(ctx context.Context, req *logical.Request) error {
	cfg, err := getConfig(ctx, req.Storage)
	if err != nil {
		return err
	}

	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return err
	}

	tag := "vault:" + pathConfigRotateRootPattern
	createAkSk, err := client.CreateAccessKey(ctx, osc.CreateAccessKeyRequest{
		Tag: &tag,
	})
	if err != nil {
		return err
	}

	oldAccessKey := cfg.AccessKey
	cfg.AccessKey = *createAkSk.AccessKey.AccessKeyId
	cfg.SecretKey = *createAkSk.AccessKey.SecretKey

	newClient, err := osc.NewClient(cfg)
	if err != nil {
		return err
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, cfg)
	if err != nil {
		return err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return err
	}

	_, err = newClient.DeleteAccessKey(ctx, osc.DeleteAccessKeyRequest{
		AccessKeyId: oldAccessKey,
	})
	if err != nil {
		b.Logger().Warn(
			pathConfigRotateRootPattern + ": failed to delete old plugin configuration access key (" + oldAccessKey + "). Error: " + err.Error(),
		)
	}

	b.setClient(newClient)

	return nil
}
