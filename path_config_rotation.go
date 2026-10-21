package outscalevault

import (
	"context"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const rootRotationStoragePath = "root-rotation-config"

const (
	rootRotationPeriodMin = 5 * time.Minute

	pathConfigRotationHelpDesc = "Controls periodic rotation of the plugin configuration access key. The rotation period must be at least five minutes and defaults to 30 days."
	pathConfigRotationHelpSyn  = "Configure automatic plugin configuration access key rotation."
	pathConfigRotationPattern  = "config/rotation"
)

type rootRotationConfig struct {
	Enabled      bool      `json:"enabled"`
	Period       int64     `json:"period"`
	NextRotation time.Time `json:"next_rotation"`
}

func pathConfigRootRotation(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: pathConfigRotationPattern,
		Fields: map[string]*framework.FieldSchema{
			"enabled": {
				Type:    framework.TypeBool,
				Default: false,
			},
			"period": {
				Type:        framework.TypeDurationSecond,
				Description: "Time between plugin configuration access key rotations. The minimum is five minutes; the default is 30 days.",
				Default:     30 * 24 * 60 * 60, // 30 days
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathRootRotationConfigRead,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback:                    b.pathRootRotationConfigUpdate,
				ForwardPerformanceSecondary: true,
				ForwardPerformanceStandby:   true,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback:                    b.pathRootRotationConfigDelete,
				ForwardPerformanceSecondary: true,
				ForwardPerformanceStandby:   true,
			},
		},
		HelpSynopsis:    pathConfigRotationHelpSyn,
		HelpDescription: pathConfigRotationHelpDesc,
	}
}

func getRootRotationConfig(ctx context.Context, storage logical.Storage) (*rootRotationConfig, error) {
	entry, err := storage.Get(ctx, rootRotationStoragePath)
	if err != nil || entry == nil {
		return nil, err
	}

	var config rootRotationConfig
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func putRootRotationConfig(
	ctx context.Context,
	storage logical.Storage,
	config *rootRotationConfig,
) error {
	entry, err := logical.StorageEntryJSON(rootRotationStoragePath, config)
	if err != nil {
		return err
	}

	return storage.Put(ctx, entry)
}

func (b *backend) pathRootRotationConfigRead(
	ctx context.Context,
	req *logical.Request,
	_ *framework.FieldData,
) (*logical.Response, error) {
	config, err := getRootRotationConfig(ctx, req.Storage)
	if err != nil || config == nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]any{
			"enabled":       config.Enabled,
			"period":        config.Period,
			"next_rotation": config.NextRotation,
		},
	}, nil
}

func (b *backend) pathRootRotationConfigUpdate(
	ctx context.Context,
	req *logical.Request,
	d *framework.FieldData,
) (*logical.Response, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	config, err := getRootRotationConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	isCreate := config == nil
	if isCreate {
		config = &rootRotationConfig{
			Enabled: d.GetDefaultOrZero("enabled").(bool),
			Period:  int64(d.GetDefaultOrZero("period").(int)),
		}
	}

	enabled := config.Enabled
	period := int(config.Period)
	if enabledRaw, ok := d.GetOk("enabled"); ok {
		enabled = enabledRaw.(bool)
	}
	if periodRaw, ok := d.GetOk("period"); ok {
		period = periodRaw.(int)
	}

	if time.Duration(period)*time.Second < rootRotationPeriodMin {
		return logical.ErrorResponse(pathConfigRotationPattern + ": period must be at least " + rootRotationPeriodMin.String()), nil
	}

	switch {
	case enabled && (isCreate || !config.Enabled || config.Period != int64(period)):
		config.NextRotation = time.Now().UTC().Add(time.Duration(period) * time.Second)
	case !enabled:
		config.NextRotation = time.Time{}
	}

	config.Enabled = enabled
	config.Period = int64(period)

	putErr := putRootRotationConfig(ctx, req.Storage, config)

	return nil, putErr
}

func (b *backend) pathRootRotationConfigDelete(
	ctx context.Context,
	req *logical.Request,
	_ *framework.FieldData,
) (*logical.Response, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	return nil, req.Storage.Delete(ctx, rootRotationStoragePath)
}

func (b *backend) periodicRootCredentialRotation(
	ctx context.Context,
	req *logical.Request,
) error {
	if !b.WriteSafeReplicationState() {
		return nil
	}
	var rotationConfig *rootRotationConfig
	var config *configEntry
	var err error

	check := func() (bool, error) {
		rotationConfig, err = getRootRotationConfig(ctx, req.Storage)
		if err != nil {
			return false, err
		}
		if rotationConfig == nil {
			b.Logger().Debug(pathConfigRotationPattern + ": rotation configuration is not set")
			return false, nil
		}

		if !rotationConfig.Enabled || time.Now().UTC().Before(rotationConfig.NextRotation) {
			return false, nil
		}

		config, err = getConfig(ctx, req.Storage)
		if err != nil {
			return false, err
		}
		if config == nil {
			b.Logger().Debug(pathConfigRotationPattern + ": plugin configuration access key is not set")
			return false, nil
		}

		return true, nil
	}

	b.lock.RLock()
	ok, err := check()
	b.lock.RUnlock()
	if err != nil || !ok {
		return err
	}

	b.lock.Lock()
	defer b.lock.Unlock()
	ok, err = check()
	if err != nil || !ok {
		return err
	}

	err = b.rotateRootCredentialLocked(ctx, req)
	if err != nil {
		return err
	}

	rotationConfig.NextRotation = time.Now().UTC().Add(time.Duration(rotationConfig.Period) * time.Second)

	err = putRootRotationConfig(ctx, req.Storage, rotationConfig)
	if err != nil {
		return err
	}

	b.Logger().Debug(pathConfigRotationPattern + ": rotation succeeded")

	return nil
}
