package outscalevault

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := Backend()
	if err := b.setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
	client     *osc.Client
	clientLock sync.Mutex
	lock       sync.RWMutex
}

func Backend() *backend {
	b := backend{}

	b.Backend = &framework.Backend{
		RunningVersion: Version,
		Help:           strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{},
			SealWrapStorage: []string{
				"config",
				"role/*",
				"root-role/*",
			},
		},
		Paths: framework.PathAppend(
			pathUserRole(&b),
			pathRootRole(&b),
			[]*framework.Path{
				pathConfig(&b),
			},
			[]*framework.Path{
				pathConfigRotateRoot(&b),
			},
			[]*framework.Path{
				pathConfigRootRotation(&b),
			},
			[]*framework.Path{
				pathUserCreds(&b),
			},
			[]*framework.Path{
				pathRootCreds(&b),
			},
		),
		Secrets: []*framework.Secret{
			secretUserAccessKeys(&b),
			secretRootAccessKeys(&b),
		},
		BackendType:      logical.TypeLogical,
		Invalidate:       b.invalidate,
		RotateCredential: b.rotateRootCredential,
		PeriodicFunc:     b.periodicRootCredentialRotation,
	}

	return &b
}

func (b *backend) setup(ctx context.Context, conf *logical.BackendConfig) error {
	return b.Setup(ctx, conf)
}

func (b *backend) reset() {
	b.clientLock.Lock()
	defer b.clientLock.Unlock()
	b.client = nil
}

func (b *backend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

func (b *backend) getClient(ctx context.Context, s logical.Storage) (*osc.Client, error) {
	b.clientLock.Lock()
	defer b.clientLock.Unlock()
	if b.client != nil {
		return b.client, nil
	}

	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errConfigNotSet
	}

	b.client, err = osc.NewClient(&config.Profile)
	if err != nil {
		return nil, err
	}

	return b.client, nil
}

func (b *backend) setClient(client *osc.Client) {
	b.clientLock.Lock()
	defer b.clientLock.Unlock()

	b.client = client
}

const backendHelp = `
The OUTSCALE secrets engine generates leased OUTSCALE access keys.

It can generate access keys for two types of identities:

- EIM user access keys: creates a temporary EIM user and an access key
  belonging to that user.
- Root user access keys: creates an access key belonging to the
  OUTSCALE root user.

The plugin requires an existing root user access key at "config" to call the
OUTSCALE API.
`
