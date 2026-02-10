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
	if err := b.setup(conf); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend
	lock   sync.RWMutex
	client *osc.Client
}

func Backend() *backend {
	b := backend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{},
			SealWrapStorage: []string{
				"config",
				"role/*",
			},
		},
		Paths: framework.PathAppend(
			pathRole(&b),
			[]*framework.Path{
				pathConfig(&b),
			},
			[]*framework.Path{
				pathCreds(&b),
			},
			[]*framework.Path{
				pathConfigRotateRoot(&b),
			},
		),
		Secrets: []*framework.Secret{
			secretAccessKeys(&b),
		},
		BackendType:      logical.TypeLogical,
		Invalidate:       b.invalidate,
		RotateCredential: b.rotateRootCredential,
	}

	return &b
}

func (*backend) setup(conf *logical.BackendConfig) error {
	// TODO: Implement configuration loading
	return nil
}

func (b *backend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.client = nil
}

func (b *backend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

func (b *backend) getClient(ctx context.Context, s logical.Storage) (*osc.Client, error) {
	b.lock.RLock()
	unlockFunc := b.lock.RUnlock
	defer func() { unlockFunc() }()

	if b.client != nil {
		return b.client, nil
	}

	b.lock.RUnlock()
	b.lock.Lock()
	unlockFunc = b.lock.Unlock

	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, err
	}

	b.client, err = osc.NewClient(config)
	if err != nil {
		return nil, err
	}

	return b.client, nil
}

const backendHelp = `
The Outsclae secretts backend is cool.
`
