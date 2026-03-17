#!/usr/bin/env bash
# Copyright IBM Corp. 2017, 2025
# SPDX-License-Identifier: MPL-2.0


set -e

CGO_ENABLED=0 GOOS=linux go build -o vault-plugin-secrets-outscale cmd/vault-plugin-secrets-outscale/main.go

docker kill vaultplg 2>/dev/null || true
docker rm vaultplg 2>/dev/null || true
tmpdir=$(mktemp -d vaultplgXXXXXX)
mkdir -p "$tmpdir/data"
chmod 777 "$tmpdir/data"
docker pull hashicorp/vault
docker run -d -p8200:8200 --name vaultplg -v "$(pwd)/$tmpdir/data":/data -v $(pwd):/example --cap-add=IPC_LOCK -e 'VAULT_LOCAL_CONFIG=
{
  "backend": {"file": {"path": "/data"}},
  "listener": [{"tcp": {"address": "0.0.0.0:8200", "tls_disable": true}}],
  "plugin_directory": "/example",
  "log_level": "debug",
  "disable_mlock": true,
  "api_addr": "http://localhost:8200"
}
' hashicorp/vault server

export VAULT_ADDR=http://localhost:8200

while ! vault status -address=http://localhost:8200 2>&1 | grep -q "^Key"; do
  sleep 1
done

initoutput=$(vault operator init -address=http://localhost:8200 -key-shares=1 -key-threshold=1 -format=json)
vault operator unseal -address=http://localhost:8200 $(echo "$initoutput" | jq -r .unseal_keys_hex[0])

export VAULT_TOKEN=$(echo "$initoutput" | jq -r .root_token)

vault plugin register -address=http://localhost:8200 -sha256=$(shasum -a 256 vault-plugin-secrets-outscale | cut -d' ' -f1) secret vault-plugin-secrets-outscale

vault secrets enable -address=http://localhost:8200 -path=outscale -plugin-name=vault-plugin-secrets-outscale plugin

vault write -address=http://localhost:8200 outscale/config \
    access_key=${OSC_ACCESS_KEY} \
    secret_key=${OSC_SECRET_KEY} \
    region=${OSC_REGION}

vault read -address=http://localhost:8200 outscale/config
