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

	if config.Protocol == "" {
		config.Protocol = "https"
	}

	if config.Region == "" {
		config.Region = "eu-west-2"
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

	return &logical.Response{
		Data: map[string]interface{}{
			"region": config.Region,
		},
	}, nil
}

func (b *backend) pathConfigWrite(
	ctx context.Context,
	req *logical.Request,
	data *framework.FieldData,
) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	createOperation := (req.Operation == logical.CreateOperation)

	if config == nil {
		if !createOperation {
			return nil, errors.New("config not found during update operation")
		}
		config = new(profile.Profile)
	}

	if access_key, ok := data.GetOk("access_key"); ok {
		config.AccessKey = access_key.(string)
	} else if !ok && createOperation {
		return nil, fmt.Errorf("missing access_key in configuration")
	}

	if secret_key, ok := data.GetOk("secret_key"); ok {
		config.SecretKey = secret_key.(string)
	} else if !ok && createOperation {
		return nil, fmt.Errorf("missing secret_key in configuration")
	}

	if region, ok := data.GetOk("region"); ok {
		config.Region = region.(string)
	}

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
			// TODO: put other fields for profile
			"access_key": {
				Type:        framework.TypeString,
				Description: "Outscale API access key",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Access key",
					Sensitive: true,
				},
			},
			"secret_key": {
				Type:        framework.TypeString,
				Description: "Outscale API secret key",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Access key",
					Sensitive: true,
				},
			},
			"region": {
				Type:        framework.TypeString,
				Description: "Outscale API region",
				Default:     "eu-west-2",
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

const pathConfigHelpSynopsis = "Configuration for Outscale backend."

const pathConfigHelpDescription = "need AK/SK, region default to eu-west-2"
