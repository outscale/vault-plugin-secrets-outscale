## Vault Outscale Secrets Plugin

[![Project Stage](https://docs.outscale.com/fr/userguide/_images/Project-Sandbox-yellow.svg)](https://docs.outscale.com/en/userguide/Open-Source-Projects.html) [![](https://dcbadge.limes.pink/api/server/HUVtY5gT6s?style=flat&theme=default-inverted)](https://discord.gg/HUVtY5gT6s)

<p align="center">
  <img alt="<Project Logo or Icon>" src="<Logo URL or Placeholder>" width="100px">
</p>

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
vault write outscale/config access_key="$OSC_ACCESS_KEY" secret_key="$OSC_SECRET_KEY"
vault read outscale/config

vault write outscale/roles/<role name> user_group_names="<group name>" max_ttl=1d ttl=10m
vault read outscale/roles/<role name>
```

### Usage

```bash
vault read outscale/creds/<role name>
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

### Notes for reuse:
- Replace all `<...>` placeholders with your content.
- You can prefill the `Project Stage` badge with values like:
  - `Project-Incubating-blue.svg`
  - `Project-Graduated-green.svg`
- You may include platform-specific instructions (macOS/Linux/Windows) in collapsible `<details>` blocks if needed.

>Labels are centrally managed in the outscale/.github repository (labels.yml).
>This repo includes a workflow (.github/workflows/sync-labels.yml) that syncs labels from there.
