package outscalevault

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/profile"
)

const (
	configStoragePath = "config"
)

var errConfigNotSet = errors.New("plugin configuration access key is not set")

func getConfig(ctx context.Context, s logical.Storage) (*profile.Profile, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	config := new(profile.Profile)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("error reading root config: %w", err)
	}

	return config, nil
}

func (b *backend) pathConfigRead(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]any{
			"region":          config.Region,
			"tls_skip_verify": config.TlsSkipVerify,
			"endpoint":        config.Endpoints.API,
			"protocol":        config.Protocol,
		},
	}, nil
}

func (b *backend) pathConfigWrite(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	createOperation := (req.Operation == logical.CreateOperation)

	if config == nil {
		if !createOperation {
			return logical.ErrorResponse("config not found during update operation"), nil
		}
		config = new(profile.Profile)
	}

	if access_key, ok := data.GetOk("access_key"); ok {
		config.AccessKey = access_key.(string)
	} else if !ok && createOperation {
		return logical.ErrorResponse("missing access_key in configuration"), nil
	}

	if secret_key, ok := data.GetOk("secret_key"); ok {
		config.SecretKey = secret_key.(string)
	} else if !ok && createOperation {
		return logical.ErrorResponse("missing secret_key in configuration"), nil
	}

	if tls_skip_verify, ok := data.GetOk("tls_skip_verify"); ok {
		config.TlsSkipVerify = tls_skip_verify.(bool)
	}

	if endpoint, ok := data.GetOk("endpoint"); ok {
		config.Endpoints.API = endpoint.(string)
	}

	config.Region = data.Get("region").(string)
	config.Protocol = data.Get("protocol").(string)

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	b.reset()

	return nil, nil
}

func (b *backend) pathConfigDelete(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	err := req.Storage.Delete(ctx, configStoragePath)

	if err == nil {
		b.reset()
	}

	return nil, err
}

func pathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"access_key": {
				Type:        framework.TypeString,
				Description: "OUTSCALE API access key of the plugin configuration. It must belong to the OUTSCALE root user.",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name: "Access key",
				},
			},
			"secret_key": {
				Type:        framework.TypeString,
				Description: "OUTSCALE API secret key of the plugin configuration.",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Secret key",
					Sensitive: true,
				},
			},
			"region": {
				Type:        framework.TypeString,
				Description: "OUTSCALE API region.",
				Default:     "eu-west-2",
			},
			"endpoint": {
				Type:        framework.TypeString,
				Description: "OUTSCALE API endpoint.",
			},
			"tls_skip_verify": {
				Type:        framework.TypeBool,
				Description: "Disable TLS certificate verification. Not recommended for production used.",
				Default:     false,
			},
			"protocol": {
				Type:        framework.TypeString,
				Description: "Protocol used for API requests. Supported values are \"http\" and \"https\".",
				Default:     "https",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},
		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    pathConfigHelpSynopsis,
		HelpDescription: pathConfigHelpDescription,
	}
}

func (b *backend) pathConfigExistenceCheck(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

const pathConfigHelpSynopsis = "Configure the plugin configuration access key."

const pathConfigHelpDescription = `Configures the plugin configuration access key and connection settings. This
access key is used by the plugin to call the OUTSCALE API. It must belong to
the OUTSCALE root user.
`
