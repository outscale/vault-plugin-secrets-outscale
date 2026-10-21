package outscalevault

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	b, storage := getTestBackend(t)

	t.Run("Create without secret key", func(t *testing.T) {
		err := testConfigCreate(t, b, storage, map[string]any{"access_key": "test-access-key"})
		require.ErrorContains(t, err, "missing secret_key in configuration")
	})
	t.Run("Update missing config", func(t *testing.T) {
		require.Error(t, testConfigUpdate(t, b, storage, map[string]any{"region": "us-east-1"}))
	})
	t.Run("Read missing config", func(t *testing.T) {
		require.NoError(t, testConfigRead(t, b, storage, nil))
	})

	entry, err := logical.StorageEntryJSON(configStoragePath, &configEntry{
		Region:    "eu-west-2",
		Protocol:  "https",
		AccountID: "test-account-id",
	})
	require.NoError(t, err)
	require.NoError(t, storage.Put(t.Context(), entry))

	t.Run("Read config", func(t *testing.T) {
		require.NoError(t, testConfigRead(t, b, storage, map[string]any{
			"account_id":      "test-account-id",
			"region":          "eu-west-2",
			"tls_skip_verify": false,
			"protocol":        "https",
		}))
	})
	t.Run("Delete config", func(t *testing.T) {
		require.NoError(t, testConfigDelete(t, b, storage))
	})
}

func TestAccConfig(t *testing.T) {
	if !runAcceptanceTests {
		t.Skip("Acceptance tests skipped unless VAULT_ACC=1 is set")
	}

	env, err := newAcceptanceTestEnv(t.Context())
	require.NoError(t, err)
	if env.AccessKey == "" || env.SecretKey == "" || env.Region == "" {
		t.Skip("OSC_ACCESS_KEY, OSC_SECRET_KEY, and OSC_REGION are required")
	}

	env.AddConfig(t)
	resp := env.ReadConfig(t)
	require.NotEmpty(t, resp.Data["account_id"])
	require.Equal(t, env.Region, resp.Data["region"])
	require.NotContains(t, resp.Data, "secret_key")
	accountID := resp.Data["account_id"]

	env.UpdateConfig(t, map[string]any{"secret_key": env.SecretKey})
	resp = env.ReadConfig(t)
	require.Equal(t, accountID, resp.Data["account_id"])
	require.Equal(t, env.Region, resp.Data["region"])

	require.Error(t, testConfigUpdate(t, env.Backend, env.Storage, map[string]any{"secret_key": "invalid"}))
	config, err := getConfig(t.Context(), env.Storage)
	require.NoError(t, err)
	require.Equal(t, env.SecretKey, config.SecretKey)
	require.Equal(t, accountID, config.AccountID)

	env.DeleteConfig(t)
}

func testConfigCreate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]any) error {
	t.Helper()
	resp, err := b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigUpdate(t *testing.T, b logical.Backend, s logical.Storage, d map[string]any) error {
	t.Helper()
	resp, err := b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "config",
		Data:      d,
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}

func testConfigRead(t *testing.T, b logical.Backend, s logical.Storage, expected map[string]any) error {
	t.Helper()
	resp, err := b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "config",
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp == nil && expected == nil {
		return nil
	}

	if resp.IsError() {
		return resp.Error()
	}

	for k, expectedV := range expected {
		actualV, ok := resp.Data[k]
		if !ok {
			assert.Failf(t, "expected key not found in response", "key %s", k)
		}
		assert.Equal(t, expectedV, actualV, "key %s mismatch", k)
	}

	return nil
}

func testConfigDelete(t *testing.T, b logical.Backend, s logical.Storage) error {
	t.Helper()
	resp, err := b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "config",
		Storage:   s,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.IsError() {
		return resp.Error()
	}
	return nil
}
