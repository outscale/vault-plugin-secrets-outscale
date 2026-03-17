package outscalevault

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

func TestRootCreds(t *testing.T) {
	b, s := getTestBackend(t)

	t.Run("Read root creds without role fails", func(t *testing.T) {
		_, err := b.HandleRequest(t.Context(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "root-creds/nonexistent",
			Storage:   s,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "role is nil")
	})
}

func TestRootCredsWithRole(t *testing.T) {
	if !runAcceptanceTests {
		t.Skip("Acceptance tests skipped unless VAULT_ACC=1 is set")
	}

	env, err := newAcceptanceTestEnv()
	require.NoError(t, err)
	require.NotNil(t, env)

	t.Cleanup(func() {
		for _, secret := range env.Secrets {
			_, err := env.Backend.HandleRequest(env.Context, &logical.Request{
				Operation: logical.RevokeOperation,
				Path:      "root-creds/cleanup",
				Storage:   env.Storage,
				Secret:    secret,
			})
			if err != nil {
				t.Logf("failed to revoke secret: %v", err)
			}
		}
	})

	t.Run("Add config", func(t *testing.T) {
		env.AddConfig(t)
	})

	t.Run("Create root role", func(t *testing.T) {
		resp, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      "root-role/" + testRoleName,
			Storage:   env.Storage,
			Data: map[string]any{
				"ttl": 3600,
			},
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Read root creds for existing role", func(t *testing.T) {
		resp, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "root-creds/" + testRoleName,
			Storage:   env.Storage,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Data["access_key"])
		require.NotEmpty(t, resp.Data["secret_key"])
		require.NotNil(t, resp.Secret)
		env.Secrets = append(env.Secrets, resp.Secret)
	})

	t.Run("Read root creds twice returns different keys", func(t *testing.T) {
		resp1, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "root-creds/" + testRoleName,
			Storage:   env.Storage,
		})
		require.NoError(t, err)
		require.NotNil(t, resp1)
		env.Secrets = append(env.Secrets, resp1.Secret)

		resp2, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "root-creds/" + testRoleName,
			Storage:   env.Storage,
		})
		require.NoError(t, err)
		require.NotNil(t, resp2)
		env.Secrets = append(env.Secrets, resp2.Secret)

		require.NotEqual(t, resp1.Data["access_key"], resp2.Data["access_key"])
		require.NotEqual(t, resp1.Data["secret_key"], resp2.Data["secret_key"])
	})
}
