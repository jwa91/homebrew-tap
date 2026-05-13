# homebrew-tap

## CLI Family Conventions

This repo is part of the jwa91 CLI tool family (agentskills, jwa-tobrew, jwa-harden, prehandover, homebrew-tap).

### Import Contract (ADR 0009)
- Process-level integration only — no Go source imports between sibling CLIs
- jwa-tobrew may exec the others; reverse is forbidden
- homebrew-tap is a pure distribution sink

### Repo Structure
- Every repo has: `.goreleaser.yaml`, `Makefile` (with `build`/`install`/`check`/`release` targets), codesign + notarize scripts, CI + release workflows, `.env.template`
- One binary per repo, name matches repo name

### GoReleaser
- All use `version: 2`, build `darwin+linux` × `amd64+arm64`, `CGO_ENABLED=0`
- Post-build codesign hook
- Modern `homebrew_casks:` block (never deprecated `brews:`)
- Archive naming: `{name}_{version}_{os}_{arch}.tar.gz`

### Distribution (ADR 0008)
- All binaries ship as Homebrew Casks (not Formulae) into `jwa91/homebrew-tap/Casks/`
- `Formula/` directory is intentionally empty

### Release Process
- `workflow_dispatch` triggers: verify → release → smoke test
- Locally: `make release VERSION=x.y.z` runs 5 preflight checks (VERSION set, 1Password signed in, gh auth'd, notarytool ready, tag at HEAD) then invokes `jwa-harden run -- goreleaser release --clean`

### Secrets (ADR 0005)
- No `.env` files ever — only `.env.template` with `op://` references resolved at runtime through 1Password

### Codesign & Notarize
- Shared `codesign.sh` and `notarize-darwin.sh` scripts
- Graceful degradation in CI when `MACOS_SIGN_IDENTITY` is unset

### CI
- Baseline workflow standardized across all repos
- Release verify gate runs before goreleaser

### Familylint
- ~35 lint rules in `jwa-tobrew`'s `internal/familylint/` enforce these conventions mechanically
- Full ADR index lives in `homebrew-tap/docs/adr/`
