package outscalevault

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
)

func pathConfigRotateRoot(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/rotate-root",
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

const (
	pathConfigRotateRootHelpDesc = "This endpoint rotates the access_key. This operation is irreversible and should be used with extreme caution."
	pathConfigRotateRootHelpSyn  = "Rotate the root access_key."
)

func (b *backend) pathConfigRotateRootUpdate(
	ctx context.Context,
	req *logical.Request,
	_ *framework.FieldData,
) (*logical.Response, error) {
	err := b.rotateRootCredential(ctx, req)
	if err != nil {
		b.Logger().Error("failed to rotate root credential", "error", err)
	}

	return nil, err
}

func (b *backend) rotateRootCredential(ctx context.Context, req *logical.Request) error {
	// Make sure we have a client available before locking the state
	client, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return err
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	cfg, err := getConfig(ctx, req.Storage)
	if err != nil {
		return err
	}

	createAkSk, err := client.CreateAccessKey(ctx, osc.CreateAccessKeyRequest{})
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

	_, err = newClient.DeleteAccessKey(ctx, osc.DeleteAccessKeyRequest{
		AccessKeyId: oldAccessKey,
	})
	if err != nil {
		return nil
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, cfg)
	if err != nil {
		return err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return err
	}

	b.client = newClient

	return nil
}
