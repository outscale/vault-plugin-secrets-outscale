## Vault Outscale Secrets Plugin

[![Project Stage](https://docs.outscale.com/fr/userguide/_images/Project-Sandbox-yellow.svg)](https://docs.outscale.com/en/userguide/Open-Source-Projects.html) [![](https://dcbadge.limes.pink/api/server/HUVtY5gT6s?style=flat&theme=default-inverted)](https://discord.gg/HUVtY5gT6s)

---

## 🌐 Links

- Documentation: <https://docs.outscale.com/en/>
- Project website: <https://github.com/outscale/vault-plugin-secrets-outscale>
- Join our community on [Discord](https://discord.gg/HUVtY5gT6s)

---

## 📄 Table of Contents

- [Overview](#-overview)
- [Requirements](#-requirements)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#usage)
- [License](#-license)
- [Contributing](#-contributing)

---

## 🧭 Overview

Vault Outscale Secrets Plugin is a Vault and OpenBao secrets engine for generating short-lived OUTSCALE credentials.

Key features:

- **EIM user access keys**: Creates a temporary EIM user and a leased access key owned by that user, with optional membership in configured EIM user groups.
- **Root user access keys**: Creates a leased access key owned by the OUTSCALE root user.
- **Plugin configuration access key rotation**: Manually or automatically rotates the plugin configuration access key stored in `<mount>/config` and used to call the OUTSCALE API.

---

## ✅ Requirements

- Vault or OpenBao
- OUTSCALE API credentials
- Go 1.25

---

## ⚙ Installation

### Option 1: Download the latest release

Download the archive for your operating system and architecture from the [Releases page](https://github.com/outscale/vault-plugin-secrets-outscale/releases).

### Option 2: Build from source

```bash
git clone https://github.com/outscale/vault-plugin-secrets-outscale.git
cd vault-plugin-secrets-outscale
```

Build with Go:

```bash
CGO_ENABLED=0 go build \
  -o vault-plugin-secrets-outscale \
  ./cmd/vault-plugin-secrets-outscale/main.go
```

Or build for the current operating system and architecture with GoReleaser:

```bash
goreleaser build \
  --clean \
  --single-target \
  --skip=validate \
  --output vault-plugin-secrets-outscale
```

Set `PLUGIN_DIR` to the `plugin_directory` configured in Vault or OpenBao.

```bash
PLUGIN_DIR=/path/to/plugin-directory
install -m 0755 vault-plugin-secrets-outscale "$PLUGIN_DIR/vault-plugin-secrets-outscale"

PLUGIN_SHA256=$(sha256sum "$PLUGIN_DIR/vault-plugin-secrets-outscale" | cut -d' ' -f1)
vault plugin register \
  -sha256="$PLUGIN_SHA256" \
  -command=vault-plugin-secrets-outscale \
  secret vault-plugin-secrets-outscale
```

---

## 🛠 Configuration

### Example

```bash
# Enable the plugin
vault secrets enable -path=outscale -plugin-name=vault-plugin-secrets-outscale plugin

# Configure the plugin configuration access key
vault write outscale/config \
  access_key="$OSC_ACCESS_KEY" \
  secret_key="$OSC_SECRET_KEY" \
  region="$OSC_REGION"

vault read outscale/config
```

### Usage

#### User Roles

A user role is a Vault configuration object that defines the temporary EIM user, its user groups, and the lease duration of its access key.

```bash
# Create a user role
vault write outscale/roles/<role-name> user_group_names="<group-name>" ttl=1h max_ttl=24h

# List user roles
vault list outscale/roles

# Read a user role
vault read outscale/roles/<role-name>

# Delete a user role
vault delete outscale/roles/<role-name>

# Generate an EIM user access key
vault read outscale/creds/<role-name>
```

#### Root Roles

A root role is a Vault configuration object that defines the lease duration of access keys owned by the OUTSCALE root user.

```bash
# Create a root role
vault write outscale/root-role/<role-name> ttl=1h

# List root roles
vault list outscale/root-role

# Read a root role
vault read outscale/root-role/<role-name>

# Delete a root role
vault delete outscale/root-role/<role-name>

# Generate a root user access key with the role TTL
vault read outscale/root-creds/<role-name>

# Generate a root user access key with a specific TTL
vault read outscale/root-creds/<role-name> ttl=10m
```

#### Plugin Configuration Access Key Rotation

This path rotates the plugin configuration access key: the root user access key stored in `<mount>/config` and used by the plugin to call the OUTSCALE API.
It does not rotate root user access keys issued through `root-creds/<role>`; those are managed through their lease and TTL.

```bash
vault write outscale/config/rotate-root
```

#### Automatic Plugin Configuration Access Key Rotation

```bash
# Enable rotation with the default 30-day period
vault write outscale/config/rotation enabled=true

# Set a custom 24-hour rotation period
vault write outscale/config/rotation period=24h

# Read rotation configuration
vault read outscale/config/rotation

# Disable rotation
vault write outscale/config/rotation enabled=false
```

---

## 📜 License

Vault Outscale Secrets Plugin is released under the BSD 3-Clause license.

© 2025 Outscale SAS

See [LICENSE](./LICENSE) for full details.

---

## 🤝 Contributing

We welcome contributions!

Please read our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before submitting a pull request.

---
