package outscalevault

import (
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAccessKey = "test-access-key"
	testSecretKey = "test-secret-key"
	testRegion    = "eu-west-2"
)

func TestConfig(t *testing.T) {
	b, storage := getTestBackend(t)

	t.Run("Create config", func(t *testing.T) {
		err := testConfigCreate(t, b, storage, map[string]any{
			"access_key": testAccessKey,
			"secret_key": testSecretKey,
			"region":     testRegion,
		})
		require.NoError(t, err)
	})

	t.Run("Read config", func(t *testing.T) {
		err := testConfigRead(t, b, storage, map[string]any{
			"region":          testRegion,
			"tls_skip_verify": false,
			"protocol":        "https",
		})
		require.NoError(t, err)
	})

	t.Run("Update config", func(t *testing.T) {
		err := testConfigUpdate(t, b, storage, map[string]any{
			"access_key": testAccessKey,
			"secret_key": testSecretKey,
			"region":     "us-east-1",
		})
		require.NoError(t, err)
	})

	t.Run("Read updated config", func(t *testing.T) {
		err := testConfigRead(t, b, storage, map[string]any{
			"region":          "us-east-1",
			"tls_skip_verify": false,
			"protocol":        "https",
		})
		require.NoError(t, err)
	})

	t.Run("Delete config", func(t *testing.T) {
		err := testConfigDelete(t, b, storage)
		require.NoError(t, err)
	})
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
