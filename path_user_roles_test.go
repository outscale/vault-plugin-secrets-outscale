package outscalevault

import (
	"strconv"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

const (
	testRoleName = "testUserRole"
	testTTL      = int64(120)
	testMaxTTL   = int64(3600)
)

func TestUserRoles(t *testing.T) {
	backend, storage := getTestBackend(t)

	t.Run("List all roles returns empty", func(t *testing.T) {
		resp, err := testRoleList(t, backend, storage)
		require.NoError(t, err)
		require.NotNil(t, resp)
		keys, ok := resp.Data["keys"].([]string)
		if ok {
			require.Empty(t, keys)
		}
	})

	t.Run("Create user role", func(t *testing.T) {
		resp, err := testUserRoleCreate(t, backend, storage, testRoleName, map[string]any{
			"user_group_names": []string{"test-group"},
			"ttl":              testTTL,
			"max_ttl":          testMaxTTL,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Read user role", func(t *testing.T) {
		resp, err := testUserRoleRead(t, backend, storage, testRoleName)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, []string{"test-group"}, resp.Data["user_group_names"])
		require.InDelta(t, float64(testTTL), resp.Data["ttl"], 0)
		require.InDelta(t, float64(testMaxTTL), resp.Data["max_ttl"], 0)
	})

	t.Run("Update user role", func(t *testing.T) {
		resp, err := testUserRoleUpdate(t, backend, storage, testRoleName, map[string]any{
			"user_group_names": []string{"updated-group"},
			"ttl":              180,
			"max_ttl":          7200,
		})
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Re-read updated user role", func(t *testing.T) {
		resp, err := testUserRoleRead(t, backend, storage, testRoleName)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, []string{"updated-group"}, resp.Data["user_group_names"])
		require.InDelta(t, float64(180), resp.Data["ttl"], 0)
		require.InDelta(t, float64(7200), resp.Data["max_ttl"], 0)
	})

	t.Run("Delete user role", func(t *testing.T) {
		resp, err := testUserRoleDelete(t, backend, storage, testRoleName)
		require.NoError(t, err)
		require.Nil(t, resp)
	})

	t.Run("Role not found after delete", func(t *testing.T) {
		resp, err := testUserRoleRead(t, backend, storage, testRoleName)
		require.NoError(t, err)
		require.Nil(t, resp)
	})
}

func TestUserRoleList(t *testing.T) {
	b, s := getTestBackend(t)

	for i := 1; i <= 5; i++ {
		roleName := testRoleName + strconv.Itoa(i)
		_, err := testUserRoleCreate(t, b, s, roleName, map[string]any{
			"user_group_names": []string{"group-" + strconv.Itoa(i)},
			"ttl":              testTTL,
			"max_ttl":          testMaxTTL,
		})
		require.NoError(t, err)
	}

	resp, err := testRoleList(t, b, s)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Data["keys"].([]string), 5)
}

func testUserRoleCreate(t *testing.T, b *backend, s logical.Storage, name string, d map[string]any) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "roles/" + name,
		Data:      d,
		Storage:   s,
	})
}

func testUserRoleUpdate(t *testing.T, b *backend, s logical.Storage, name string, d map[string]any) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/" + name,
		Data:      d,
		Storage:   s,
	})
}

func testUserRoleRead(t *testing.T, b *backend, s logical.Storage, name string) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/" + name,
		Storage:   s,
	})
}

func testUserRoleDelete(t *testing.T, b *backend, s logical.Storage, name string) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "roles/" + name,
		Storage:   s,
	})
}

func testRoleList(t *testing.T, b *backend, s logical.Storage) (*logical.Response, error) {
	t.Helper()
	return b.HandleRequest(t.Context(), &logical.Request{
		Operation: logical.ListOperation,
		Path:      "roles/",
		Storage:   s,
	})
}
