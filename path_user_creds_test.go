package outscalevault

import (
	"context"
	"testing"

	"github.com/hashicorp/vault/sdk/helper/template"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
	"github.com/stretchr/testify/require"
)

func TestUserCreds(t *testing.T) {
	b, s := getTestBackend(t)

	t.Run("Read creds without role fails", func(t *testing.T) {
		resp, err := b.HandleRequest(t.Context(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "creds/nonexistent",
			Storage:   s,
		})
		require.NoError(t, err)
		require.True(t, resp.IsError())
		require.Contains(t, resp.Error().Error(), "role \"nonexistent\" does not exist")
	})
}

func TestAccUserCredsWithRole(t *testing.T) {
	if !runAcceptanceTests {
		t.Skip("Acceptance tests skipped unless VAULT_ACC=1 is set")
	}

	env, err := newAcceptanceTestEnv(t.Context())
	require.NoError(t, err)
	require.NotNil(t, env)

	t.Run("Add config", func(t *testing.T) {
		env.AddConfig(t)
	})
	client, err := env.Backend.(*backend).getClient(env.Context, env.Storage)
	require.NoError(t, err)

	groupTemplate, err := template.NewTemplate(template.Template("vault-test-{{random 20}}"))
	require.NoError(t, err)
	groupName, err := groupTemplate.Generate(nil)
	require.NoError(t, err)
	_, err = client.CreateUserGroup(env.Context, osc.CreateUserGroupRequest{UserGroupName: groupName})
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())
		for _, secret := range env.Secrets {
			_, err := env.Backend.HandleRequest(ctx, &logical.Request{
				Operation: logical.RevokeOperation,
				Path:      "creds/cleanup",
				Storage:   env.Storage,
				Secret:    secret,
			})
			if err != nil {
				t.Logf("failed to revoke secret: %v", err)
			}
		}
		_, err := client.DeleteUserGroup(ctx, osc.DeleteUserGroupRequest{UserGroupName: groupName})
		if err != nil {
			t.Errorf("failed to delete test user group: %v", err)
		}
	})

	t.Run("Create user role", func(t *testing.T) {
		resp, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.CreateOperation,
			Path:      "roles/" + testRoleName,
			Storage:   env.Storage,
			Data: map[string]any{
				"user_group_names": []string{groupName},
				"ttl":              3600,
				"max_ttl":          7200,
			},
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Read creds for existing role", func(t *testing.T) {
		resp, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "creds/" + testRoleName,
			Storage:   env.Storage,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Data["access_key"])
		require.NotEmpty(t, resp.Data["secret_key"])
		require.NotNil(t, resp.Secret)
		env.Secrets = append(env.Secrets, resp.Secret)
	})

	t.Run("Read creds twice returns different keys", func(t *testing.T) {
		resp1, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "creds/" + testRoleName,
			Storage:   env.Storage,
		})
		require.NoError(t, err)
		require.NotNil(t, resp1)
		env.Secrets = append(env.Secrets, resp1.Secret)

		resp2, err := env.Backend.HandleRequest(env.Context, &logical.Request{
			Operation: logical.ReadOperation,
			Path:      "creds/" + testRoleName,
			Storage:   env.Storage,
		})
		require.NoError(t, err)
		require.NotNil(t, resp2)
		env.Secrets = append(env.Secrets, resp2.Secret)

		require.NotEqual(t, resp1.Data["access_key"], resp2.Data["access_key"])
		require.NotEqual(t, resp1.Data["secret_key"], resp2.Data["secret_key"])
	})
}
