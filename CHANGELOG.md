# 📜 Changelog

All notable changes to this project will be documented in this file.  
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)  
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### 💥 Breaking
- (nothing yet)

### ✨ Added
- (nothing yet)

### 🛠️ Changed / Refactoring
- (nothing yet)

### 📝 Documentation
- (nothing yet)

### ⚰️ Deprecated
- (nothing yet)

### 🗑️ Removed
- (nothing yet)

### 🐛 Fixed
- (nothing yet)

### 🔒 Security
- (nothing yet)

### 📦 Dependency updates
- (nothing yet)

### 🌱 Others
- (nothing yet)

---

## [0.0.2] - 2026-10-27

### 🛠️ Changed / Refactoring
* 💚 ci: add ghcr publishing & cosign signing by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/29
### 📝 Documentation
* 📝 docs: vault helpers & readme by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/25
### 📦 Dependency updates
* chore: Configure Renovate by @Open-Source-Bot in https://github.com/outscale/vault-plugin-secrets-outscale/pull/11
* ⬆️ deps(gomod): update module github.com/hashicorp/vault/api to v1.23.0 by @Open-Source-Bot in https://github.com/outscale/vault-plugin-secrets-outscale/pull/19

## [0.0.1] - 2026-10-05

### ✨ Added
* ✨ feat(root-credentials): add periodic rotation by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/6
### 🛠️ Changed / Refactoring
* ♻ refactor: convert errors to logical.ErrorResponse by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/7
### 🌱 Others
* 🎉 begin: add poc for vault secret plugin by @jobs62 in https://github.com/outscale/vault-plugin-secrets-outscale/pull/1
* ✨ feat(auth): add root credential rotation endpoint by @jobs62 in https://github.com/outscale/vault-plugin-secrets-outscale/pull/2
* 💚 chore(ci): disable CGO by @jobs62 in https://github.com/outscale/vault-plugin-secrets-outscale/pull/3
* ✨ feat: root-creds provisioning by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/4
* 💚 fix: fix gorealeser configuration by @jobs62 in https://github.com/outscale/vault-plugin-secrets-outscale/pull/5
* 💚 goreleaser: remove deprecated value by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/8
* 💚 goreleaser: bump action by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/9
* 💚 goreleaser: change token by @ryohkhn in https://github.com/outscale/vault-plugin-secrets-outscale/pull/10

---

## 🔑 How to use this file

1. During development:
   - Add entries under **[Unreleased]** in the right category.
   - Keep the wording concise and clear.
   - Link related issues/PRs where relevant (`#123`, [PR-456](https://github.com/org/repo/pull/456)).

2. When preparing a release:
   - Move items from **[Unreleased]** into a new version section (`[X.Y.Z] - YYYY-MM-DD`).
   - Leave **[Unreleased]** empty for future changes.

3. At release time:
   - GitHub automatically generates release notes from PR labels (`.github/release.yml`).
   - This `CHANGELOG.md` is the **permanent record** in the repository.
