## Vault Outscale Secrets Plugin

[![Project Stage](https://docs.outscale.com/fr/userguide/_images/Project-Sandbox-yellow.svg)](https://docs.outscale.com/en/userguide/Open-Source-Projects.html) [![](https://dcbadge.limes.pink/api/server/HUVtY5gT6s?style=flat&theme=default-inverted)](https://discord.gg/HUVtY5gT6s)

---

## 🌐 Links

- Documentation: <https://docs.outscale.com/en/>
- Project website: <https://github.com/outscale/<project-name>>
- Join our community on [Discord](https://discord.gg/HUVtY5gT6s)
- Related tools or community: <<https://example.com>> *(optional)*

---

## 📄 Table of Contents

- [Overview](#-overview)
- [Requirements](#-requirements)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [Examples](#-examples)
- [License](#-license)
- [Contributing](#-contributing)

---

## 🧭 Overview

Vault Outscale Secrets Plugin is a plugin to dynamically provision EIM users and AK/SK.

Key features:

---

## ✅ Requirements

- Go 1.25
- Access to the OUTSCALE API (with appropriate credentials)

---

## ⚙ Installation

### Option 1: Download from Releases

Download the latest binary from the [Releases page](https://github.com/outscale/vault-plugin-secrets-outscale/releases).

### Option 2: Install from source

```bash
git clone https://github.com/outscale/vault-plugin-secrets-outscale.git
cd vault-plugin-secrets-outscale
go build -o <vault install dir>/plugins/vault-plugin-secrets-outscale cmd/main.go
vault plugin register -sha256=<hash of the binnary> secret vault-plugin-secrets-outscale
````

---

## 🛠 Configuration

### Example

```bash
# Enable the plugin
vault secrets enable -path=outscale -plugin-name=vault-plugin-secrets-outscale plugin

# Configure credentials
vault write outscale/config \
    access_key="$OSC_ACCESS_KEY" \
    secret_key="$OSC_SECRET_KEY" \
    region="eu-west-2"

vault read outscale/config
```

### Usage

#### User Roles

```bash
# Create a user role
vault write outscale/roles/<role-name> user_group_names="<group-name>" ttl=1h max_ttl=24h

# List user roles
vault list outscale/roles

# Read user role
vault read outscale/roles/<role-name>

# Delete user role
vault delete outscale/roles/<role-name>

# Generate credentials
vault read outscale/creds/<role-name>
```

#### Root Roles

```bash
# Create a root role
vault write outscale/root-role/<role-name> ttl=1h

# List root roles
vault list outscale/root-role

# Read root role
vault read outscale/root-role/<role-name>

# Delete root role
vault delete outscale/root-role/<role-name>

# Generate credentials with rool-role TTL
vault read outscale/root-creds/<role-name>

# Generate credentials with specific TTL
vault read outscale/root-creds/<role-name> ttl=10min
```

#### Rotate Root Credentials

```bash
vault write outscale/config/rotate-root
```

---

## 📜 License

Vault Outscale Secrets Plugin is released under the < License Name > license.

© 2025 Outscale SAS

See [LICENSE](./LICENSE) for full details.

---

## 🤝 Contributing

We welcome contributions!

Please read our [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md) before submitting a pull request.

---
