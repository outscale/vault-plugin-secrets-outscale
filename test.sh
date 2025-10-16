#!/bin/bash

set -ex

go build -o vault/plugins/vault-plugin-secrets-outscale cmd/main.go

vault server -dev -dev-root-token-id=root -dev-plugin-dir=./vault/plugins/ -log-level=debug &
export PID=$!
function clean_kill {
    vault lease revoke -prefix outscale/

    sleep 2
	kill $PID
}
trap clean_kill EXIT

export VAULT_TOKEN="root"
export VAULT_ADDR='http://127.0.0.1:8200'

sleep 1

vault secrets enable -path=outscale vault-plugin-secrets-outscale
vault plugin info secret vault-plugin-secrets-outscale

sleep 1
vault write outscale/config access_key="$OSC_ACCESS_KEY" secret_key="$OSC_SECRET_KEY"
vault read outscale/config

sleep 1
vault write outscale/roles/test user_group_names="vault-group" max_ttl=1d ttl=10m
vault read outscale/roles/test

sleep 1
vault read outscale/creds/test

sleep 3
