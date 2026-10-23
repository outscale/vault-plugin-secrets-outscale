//go:build oci

package docker_test

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
)

const (
	pluginRepository = "vault-plugin-secrets-outscale"
	pluginName       = "vault-plugin-secrets-outscale"

	vaultAddr  = "http://127.0.0.1:8200"
	vaultToken = "root"
)

func TestOCIPluginMount(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	pluginImage := pluginRepository + ":test-" + runtime.GOARCH

	dockerClient, err := client.New(client.FromEnv)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dockerClient.Close())
	})

	image, err := dockerClient.ImageInspect(ctx, pluginImage)
	require.NoError(t, err, "plugin not foud")
	require.NotEmpty(t, image.ID)

	config := api.DefaultConfig()
	config.Address = vaultAddr

	vaultClient, err := api.NewClient(config)
	require.NoError(t, err)

	vaultClient.SetToken(vaultToken)

	err = vaultClient.Sys().RegisterPluginRuntime(ctx, &api.RegisterPluginRuntimeInput{
		Name:       "docker",
		Type:       api.PluginRuntimeTypeContainer,
		OCIRuntime: "runc",
	})
	require.NoError(t, err)

	err = vaultClient.Sys().RegisterPluginWithContext(ctx, &api.RegisterPluginInput{
		Name:     pluginName,
		Type:     api.PluginTypeSecrets,
		OCIImage: pluginRepository,
		SHA256:   strings.TrimPrefix(image.ID, "sha256:"),
		Runtime:  "docker",
	})
	require.NoError(t, err)

	err = vaultClient.Sys().MountWithContext(ctx, "outscale", &api.MountInput{
		Type: pluginName,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := vaultClient.Sys().UnmountWithContext(cleanupCtx, "outscale"); err != nil {
			t.Logf("failed to unmount test plugin: %v", err)
		}
	})

	path := "outscale/root-role/oci-test"
	_, err = vaultClient.Logical().WriteWithContext(ctx, path, map[string]any{"ttl": 60})
	require.NoError(t, err)

	secret, err := vaultClient.Logical().ReadWithContext(ctx, path)
	require.NoError(t, err)
	require.NotNil(t, secret)
	require.Equal(t, json.Number("60"), secret.Data["ttl"])
}
