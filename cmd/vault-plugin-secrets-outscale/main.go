package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	oscVault "github.com/outscale/vault-plugin-secrets-outscale"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	logger := hclog.New(&hclog.LoggerOptions{})

	flags := apiClientMeta.FlagSet()
	err := flags.Parse(os.Args[1:])
	if err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err = plugin.ServeMultiplex(&plugin.ServeOpts{
		BackendFactoryFunc: oscVault.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
