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

- **ADR 0009** — CLI family import contract. Defines the acyclic
  bootstrap order and process-boundary-only relationships between
  `jwa-harden`, `agentskills`, `prehandover`, and `jwa-tobrew`.

### Changed

- **ADR 0006 / 0007** refreshed to match the current Cask-based
  distribution model and dotfiles Brewfile entries.
- **Release skill ownership clarified**: project release skills are installed
  from `agentskills`, while `jwa-tobrew init` owns only release config
  scaffolding and process-level calls to `agentskills`.

## [0.4.1] — 2026-05-13

This release completes ADR 0008's migration: every Go CLI in the family
now ships as a **Cask** (modern `homebrew_casks:` shape) with Developer
ID signing + notarization. The deprecated `brews:` block is no longer
used anywhere in the family.

### Added

- **`Casks/agentskills.rb`** — auto-committed by `jwa91/agentskills`
  at v0.1.1 (now codesigned + notarized).
- **`Casks/prehandover.rb`** — auto-committed by `jwa91/prehandover`
  at v0.1.1 (now codesigned + notarized).

### Removed

- **`Formula/agentskills.rb`** — replaced by the Cask above.
- **`Formula/prehandover.rb`** — replaced by the Cask above.

### Changed

- **README "How items get here"** table updated — every row now points
  at `Casks/` (the Formula column is gone).
- **README Layout** notes `Formula/` is empty by design as of this
  release; future items in the family default to Casks.
- **`tap.toml`** regenerated to reflect the new Cask shapes.

### Known issues

- `jwa-tobrew config` still renders multi-line cask `depends_on` blocks
  (e.g. `depends_on formula: [\n  "gh",\n]`) as the truncated string
  `deps = ["formula: ["]` in `tap.toml`. Affects `agentskills` and
  `jwa-tobrew` entries. Cosmetic — the underlying `.rb` files are
  correct. Fix queued for the next `jwa-tobrew` release.

## [0.4.0] — 2026-05-13

This release converts the tap into a **pure distribution channel** (ADR
0008). No more Go source, no more in-tree `.goreleaser.yaml`, no more
`make bootstrap`. Sibling repos own their own pipelines and auto-commit
their `.rb` files here.

### Added

- **ADR 0008** — one binary per repo; ship via `homebrew_casks`
  (modern GoReleaser key replacing the deprecated `brews:`). Documents
  the namespace separation that resolves the v0.2.0 collision and the
  shift from Formula to Cask for pre-compiled binaries.
- **`Casks/jwa-tobrew.rb`** — auto-committed by `jwa91/jwa-tobrew` at
  v0.4.0, the first cask-shaped release of the publisher CLI.
- **`Casks/jwa-harden.rb`** — auto-committed by `jwa91/jwa-harden` at
  v0.1.0. New `jwa-*` family member; wraps a command with `op run`
  against the nearest `.env.template`. Replaces the long-standing
  "(when that exists)" placeholder in `jwa-tobrew doctor` and the
  README.

### Changed

- **Tap role** — narrowed to `Casks/`, `Formula/`, `tap.toml`, ADRs,
  `.agents/skills/` (gitignored), and tap-meta CHANGELOG/README.
- **README intro** rewritten to describe the new distribution-channel
  role and link out to the publisher repos.

### Removed

- **`tools/jwa-tobrew/`** — entire CLI source tree moved to
  `jwa91/jwa-tobrew`. History prior to v0.3.0 remains here in git
  history; new commits go to the dedicated repo.
- **`Formula/jwa-tobrew.rb`** — replaced by `Casks/jwa-tobrew.rb`.
  End-user `brew install jwa91/tap/jwa-tobrew` is unchanged
  (modern brew auto-detects cask vs formula in the tap).
- **`.goreleaser.yaml`** — was used to release the in-tree
  `jwa-tobrew`; no longer needed.
- **`go.mod`** — no Go code remains in the tap.
- **Makefile build/install/dev/bootstrap/clean targets** — gone.
  Only `bootstrap-skills`, `setup`, and `hooks` remain.
- **prek `go-fmt` / `go-vet` / `go-build` hooks** — gone (no Go code
  to lint). The `jwa-tobrew-align` hook now requires the
  brew-installed binary; the `go run` fallback was removed.

### Fixed

- **ADR 0007** updated to reflect that the tap no longer owns the
  `jwa-tobrew` publisher source — it owns only the `.rb` files,
  `tap.toml`, and ADRs.

### Known issues

- `jwa-tobrew config` parses multi-line cask `depends_on` blocks (e.g.
  `depends_on formula: [\n  "gh",\n]`) as a single truncated string —
  in `tap.toml` this appears as `deps = ["formula: ["]`. Cosmetic only;
  the underlying `Casks/*.rb` is correct. Fix queued for the next
  jwa-tobrew release.

## [0.3.0] — 2026-05-13

### Added

- **ADR 0006** (agentskills boundary contract) and **ADR 0007**
  (dotfiles contract) documenting the one-way file/install flows
  between this tap and its sibling repos.
- **Tap-skill migration plan** at
  `docs/migration-tap-skills-to-agentskills.md` covering pre-migration
  gates, file moves, planned `align.go` changes, and the SHA-pinning
  decision for `AGENTSKILLS_REF`.
- **`align` (tap mode) checks**: ADR 0006 and ADR 0007 must exist;
  no `Brewfile` at the tap root (ADR 0007 forbidden pattern (c)); no
  `security-ground-rules.md` duplicated inside the tap (ADR 0007
  reference-loop rule).
- **`prek` `jwa-tobrew-align` hook scope** extended to include
  `docs/migration-` so edits to the migration plan re-run align.
- **Three new agent skills** at `.agents/skills/` (with `.claude/skills/`
  symlinks):
  - `tap-alignment` — explains the three-layer convention model
    (ADRs / `align` rules / prek hooks), the two `align` modes, and
    how to add a new convention.
  - `scaffold-cli` — end-to-end walk-through for setting up a new Go
    CLI publishable via this tap (repo create → `init` → first
    release locally and via tag-driven CI).
  - `tap-scaffolding` — maintainer-oriented: how the embedded
    templates are wired, the `init`↔`align` contract, and the recipe
    for adding a new scaffold kind.
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
- **`tools/jwa-tobrew/manifest_test.go`** — locks in the
  deps-serialization fix with six representative cases.

### Changed

- **Migrate 4 tap-managing skills (jwa-tobrew, scaffold-cli,
  tap-alignment, tap-scaffolding) to jwa91/agentskills.** Now installed
  via `make bootstrap-skills` (agentskills bootstrap). The skill source
  of truth lives in jwa91/agentskills per ADR 0006; here they are
  generated artifacts.
- **Add `bootstrap-skills` Makefile target** pinned to
  `agentskills@beb2952`. Also adds a `setup` aggregate (`hooks` +
  `bootstrap-skills`) for fresh clones.
- **Update align.go** — remove the jwa-tobrew skill-symlink finding
  (the dir is now generated, not vendored), and add an
  agentskills-on-PATH check that surfaces a manual install hint when
  the CLI is missing.
- **Add `.agents/skills/` and `.claude/skills/` to .gitignore** (tap
  root only, leading-slash-anchored). Both trees are populated by
  `make bootstrap-skills`.
- **Tag scheme for the agentskills CLI** is plain `v*` (was `cli/v*`);
  per-skill release tags remain `<skill>/v*`. ADR 0006 and the
  migration plan updated accordingly.
- **ADR 0005** rewritten to a tap-scoped policy; the universal `.env`
  / `op://` rule moves upstream to `dotfiles/docs/adr/0001`.
- **`jwa-tobrew` skill** rewritten for the current command surface
  (adds `align`; removes outdated step-by-step that pre-dated the
  unified Go release pipeline). Calls out that the
  `.env.template` + `op://` pattern is the *current bridge* until
  `jwa-harden` lands, not a permanent contract.
- **Embedded `skill-release.md.tmpl`** updated for Go projects to
  use `goreleaser release --clean` directly (was still pointing at
  the deleted `./scripts/release.sh` for Go).
- **Unified Go release pipeline** — Go CLIs now use `.goreleaser.yaml`
  exclusively. `init --kind=go` no longer scaffolds
  `scripts/release.sh`; the local release command is
  `goreleaser release --clean` (wrapped by `op run`), matching the
  CI tag-driven flow. align enforces this: Go project missing
  `.goreleaser.yaml` → fail; cask/formula missing
  `scripts/release.sh` → fail. Closes the false positive previously
  reported against repos like `prehandover` that publish to this tap
  via their own GoReleaser `brews:` block.
- **Renamed** `templates/release-cask.sh.tmpl` → `templates/release.sh.tmpl`
  (used for both cask and formula). Removed `templates/release-go.sh.tmpl`
  entirely. Both old templates also incorrectly referenced `brewtap` —
  the surviving template now uses `jwa-tobrew`.
- **align (tap mode)** now requires `.goreleaser.yaml` at the tap root
  if the tap embeds a Go CLI (i.e. has `go.mod`).
- **`make bootstrap`** refuses if `v$(VERSION)` already exists at a
  SHA other than HEAD. The previous `|| true` swallowed the
  tag-exists error and would have silently shipped an outdated
  commit.
- **README Commands block** documents `deps` and `align`; corrects
  `config` to mention the items-table regeneration; clarifies
  `upgrade`.

### Fixed

- **`renderTapTOML` / `renderTapLocalTOML`** end on a single newline
  instead of two, so generators are idempotent under
  end-of-file-fixer (no churn loop with prek).
- **gofmt drift in `ui.go` and `github.go`** — caught by the new
  pre-commit hook on its first run.
- **`jwa-tobrew config` deps serialization** — bare formula deps
  (`depends_on "git"`) now serialize as `deps = ["git"]`, not
  `["\"git\""]`. Cask-style key:value deps unchanged.
- **`jwa-tobrew config` reports README write** — the items-table
  update has been wired in since v0.2.0 but `config` printed only
  two `✓` lines; it now reports the third write explicitly so a
  failing README path is visible.
- **Stale legacy binaries** — `bin/brewtap` and `bin/brwtp` removed
  (leftovers from the pre-`jwa-tobrew` rename).

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

[Unreleased]: https://github.com/jwa91/homebrew-tap/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/jwa91/homebrew-tap/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/jwa91/homebrew-tap/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/jwa91/homebrew-tap/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jwa91/homebrew-tap/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jwa91/homebrew-tap/releases/tag/v0.1.0
