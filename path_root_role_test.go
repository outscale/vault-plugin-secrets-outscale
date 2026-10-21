package outscalevault

import (
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

const (
	testRootRoleName = "testRootRole"
	testRootTTL      = int64(7200)
)

func TestRootRoles(t *testing.T) {
	b, s := getTestBackend(t)

	t.Run("List all root roles returns empty", func(t *testing.T) {
		resp, err := testRootRoleList(t, b, s)
		require.NoError(t, err)
		require.NotNil(t, resp)
		keys, ok := resp.Data["keys"].([]string)
		if ok {
			require.Empty(t, keys)
		}
	})

	t.Run("Create root role", func(t *testing.T) {
		resp, err := testRootRoleCreate(t, b, s, testRootRoleName, map[string]any{
			"name": testRootRoleName,
			"ttl":  testRootTTL,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Read root role", func(t *testing.T) {
		resp, err := testRootRoleRead(t, b, s, testRootRoleName)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.InDelta(t, float64(testRootTTL), resp.Data["ttl"], 0)
	})

	t.Run("Update root role", func(t *testing.T) {
		resp, err := testRootRoleUpdate(t, b, s, testRootRoleName, map[string]any{
			"ttl": 3600,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Re-read updated root role", func(t *testing.T) {
		resp, err := testRootRoleRead(t, b, s, testRootRoleName)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.InDelta(t, float64(3600), resp.Data["ttl"], 0)
	})

	t.Run("Delete root role", func(t *testing.T) {
		resp, err := testRootRoleDelete(t, b, s, testRootRoleName)
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Root role not found after delete", func(t *testing.T) {
		resp, err := testRootRoleRead(t, b, s, testRootRoleName)
		require.NoError(t, err)
		require.Nil(t, resp)
	})
}

func TestRootRoleList(t *testing.T) {
	b, s := getTestBackend(t)

	for i := 1; i <= 3; i++ {
		roleName := testRootRoleName + strconv.Itoa(i)
		_, err := testRootRoleCreate(t, b, s, roleName, map[string]any{
			"name": roleName,
			"ttl":  testRootTTL,
		})
		require.NoError(t, err)
	}

	resp, err := testRootRoleList(t, b, s)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Data["keys"].([]string), 3)
}

func TestRootRoleDefaultTTL(t *testing.T) {
	b, s := getTestBackend(t)

	t.Run("Create root role with default TTL", func(t *testing.T) {
		resp, err := testRootRoleCreate(t, b, s, "default-ttl-role", map[string]any{
			"name": "default-ttl-role",
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Read root role has default TTL", func(t *testing.T) {
		resp, err := testRootRoleRead(t, b, s, "default-ttl-role")
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.InDelta(t, float64(7200), resp.Data["ttl"], 0)
	})
}

func testRootRoleCreate(t *testing.T, b *backend, s logical.Storage, name string, d map[string]any) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "root-role/" + name,
		Data:      d,
		Storage:   s,
	})
}

func testRootRoleUpdate(t *testing.T, b *backend, s logical.Storage, name string, d map[string]any) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "root-role/" + name,
		Data:      d,
		Storage:   s,
	})
}

func testRootRoleRead(t *testing.T, b *backend, s logical.Storage, name string) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "root-role/" + name,
		Storage:   s,
	})
}

func testRootRoleDelete(t *testing.T, b *backend, s logical.Storage, name string) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "root-role/" + name,
		Storage:   s,
	})
}

func testRootRoleList(t *testing.T, b *backend, s logical.Storage) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.ListOperation,
		Path:      "root-role/",
		Storage:   s,
	})
}
