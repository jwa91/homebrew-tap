# Changelog

Versions track the evolution of this tap as a system: its structure,
conventions, the `jwa-tobrew` CLI's behaviour, and the security model.
The versions of individual items shipped by the tap (see `Casks/` and
`Formula/`) are tracked separately in their `.rb` files.

Format follows [Keep a Changelog](https://keepachangelog.com); versions
follow [SemVer](https://semver.org) interpreted for a tap repo:

- **Major** — breaking change to the CLI's command surface or to the
  schema of generated artifacts (`tap.toml`, scaffolds written by
  `init`).
- **Minor** — new commands, new conventions, or new generated files.
- **Patch** — bug fixes, doc updates, internal refactors that don't
  change outputs.

## [Unreleased]

### Added

- **`prek.toml`** — pre-commit hook config in prek's native TOML format
  (works with `pre-commit` too via `.pre-commit-config.yaml`-equivalent
  semantics). Enforces locally: gofmt + vet + build, gitleaks secret
  scan, forbid raw `.env*` files (only `.env.template` allowed),
  `jwa-tobrew align` must report no drift, conventional-commit subject
  on the commit-msg stage, plus standard whitespace/syntax/large-file
  guards. Install with `make hooks`.
- **`.github/CODEOWNERS`** — template that auto-requests review from
  the owner on every PR. Becomes enforcing once branch protection's
  "Require review from Code Owners" is enabled.
- **`make hooks`** target installs prek's git shims.

### Fixed

- **`renderTapTOML` / `renderTapLocalTOML`** end on a single newline
  instead of two, so generators are idempotent under
  end-of-file-fixer (no churn loop with prek).
- **gofmt drift in `ui.go` and `github.go`** — caught by the new
  pre-commit hook on its first run.

## [0.2.0] — 2026-05-13

### Changed

- **Folder layout** — CLI source moved from `cmd/jwa-tobrew/` to
  `tools/jwa-tobrew/` so the repo's three top-level areas (`Casks/`,
  `Formula/`, `tools/`) are visually balanced and the Go-internal
  convention isn't load-bearing for non-Go readers.
- **Skills layout** — `.claude/skills/jwa-tobrew/` migrated to
  `.agents/skills/jwa-tobrew/` with `.claude/skills/jwa-tobrew` as a
  symlink, mirroring prehandover. Lets non-Claude agent runtimes
  (codex, cursor) symlink the same source. Convention enforced by
  `jwa-tobrew align`.

### Added

- **Auto-updating README items table** — every `add` / `bump` /
  `release` regenerates the items section between `<!-- BEGIN ITEMS -->`
  and `<!-- END ITEMS -->` markers in `README.md`. Each item links to
  its source repo. README is now committed alongside the `.rb` and
  `tap.toml` in every item-touching commit.
- **`align` command** — reports (or `--apply`s) drift from the
  conventions `jwa-tobrew` expects. Auto-detects tap mode vs project
  mode from cwd contents.
- **Multi-platform formula safety** — `bump` and `release`'s update
  path refuse `.rb` files with more than one `url` line (e.g.
  GoReleaser-published formulas like `prehandover.rb`). Source repos
  own their own multi-platform updates.
- **CHANGELOG.md** (this file) — versioned history of the tap itself.

### Fixed

- **`bump` no longer skips manifest regeneration.** Previously,
  `tap.toml`'s `version` field could go stale because only the `.rb`
  file was rewritten. `bump` now calls `writeManifests` like `add`
  and `release` already did.

## [0.1.0] — 2026-05-12

### Added

- **`jwa-tobrew` CLI** with commands `add`, `bump`, `config`, `deps`,
  `doctor`, `init`, `release`, `upgrade`. Tags + GitHub-releases +
  updates the tap in one shot for Go projects (via GoReleaser), Casks
  (macOS apps), and Formulae (other binaries).
- **`tap.toml` + `tap.local.toml`** — committed manifest of every item
  with name/kind/source/version/asset/desc, plus a gitignored sidecar
  annotating each item with whether the source repo is locally cloned.
- **Architecture Decision Records** at `docs/adr/`:
  - `0001` — jwa-tobrew never touches secrets directly; reads
    `$GITHUB_TOKEN` from env, leaves storage/rotation to `jwa-harden`.
  - `0002` — Tap pushes go over SSH only; HTTPS-with-token rejected.
  - `0003` — No hardcoded secret references (UUIDs, op:// items)
    in source.
  - `0004` — Two distinct GitHub-ref parsers (single-string CLI input
    vs body scan in a `.rb`); intentional.
  - `0005` — Only `.env.template` (op:// references) is checked in;
    real `.env*` files are gitignored and CI-blocked.
- **`init` writes scaffolds** into target projects: `.goreleaser.yaml`
  (Go only), `scripts/release.sh`, `.env.template`, `.gitignore`
  entries, plus a project-scoped release skill at
  `.claude/skills/release/SKILL.md`.
- **Doctor** (`jwa-tobrew doctor`) checks PATH, tap location, SSH
  origin, `$GITHUB_TOKEN` presence.
- **Self-hosting GoReleaser config** at `.goreleaser.yaml` writes
  `Formula/jwa-tobrew.rb` back into this same repo on release.

[Unreleased]: https://github.com/jwa91/homebrew-tap/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/jwa91/homebrew-tap/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jwa91/homebrew-tap/releases/tag/v0.1.0
