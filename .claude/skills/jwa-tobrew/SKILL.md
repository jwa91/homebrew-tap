---
name: jwa-tobrew
description: Wire the current project up for releases via a personal Homebrew tap, using the `jwa-tobrew` CLI. Use when the user says "add jwa-tobrew", "publish to my homebrew tap", "make this installable via brew", or asks how to release this project to their personal tap. Handles Go binaries (via GoReleaser), macOS apps (Casks), and other binaries (Formulae).
---

# Add jwa-tobrew to a project

`jwa-tobrew` publishes a project to the user's personal Homebrew tap end-to-end: tag, GitHub Release with auto-generated changelog, asset upload, tap update + commit + push. Source and full docs: <https://github.com/jwa91/homebrew-tap>.

## Security model — important

`jwa-tobrew` is one of the `jwa-*` family of personal CLIs. **It does not touch 1Password directly.** It expects `$GITHUB_TOKEN` in its environment when it hits the GitHub API and errors clearly otherwise. The canonical way to provide that env is to wrap the invocation with `jwa-harden run` (when that CLI is available) or `op run --env-file=.env.template --` for now.

The full security model lives at `~/dotfiles/docs/security-ground-rules.md` — read it before suggesting any token-handling changes.

## Step 1 — verify prerequisites

```bash
jwa-tobrew doctor
```

Checks PATH (`gh`, `git`, `goreleaser`), tap location, that the tap origin is SSH, and that `$GITHUB_TOKEN` is reachable. Follow any printed hints before continuing.

If `jwa-tobrew` isn't installed:

```bash
brew install jwa91/tap/jwa-tobrew
```

(If the tap isn't bootstrapped on this machine yet, follow the bootstrap section in the tap repo's README.)

## Step 2 — scaffold the project

From inside the project's git repo:

```bash
jwa-tobrew init
```

Auto-detects:

- **Go** (`go.mod` present) → writes `.goreleaser.yaml` and `scripts/release.sh`. Releases use GoReleaser.
- **macOS app** (`*.xcodeproj`, `*.xcworkspace`, or `Package.swift`) → writes `scripts/release.sh` for Cask publishing.
- **Other** → pass `--kind formula` and provide the binary path at release time.

Useful flags: `--name`, `--desc`, `--kind go|cask|formula`, `--force`.

## Step 3 — first release

```bash
op run --env-file=.env.template -- ./scripts/release.sh 0.1.0                       # Go
op run --env-file=.env.template -- ./scripts/release.sh 0.1.0 build/MyApp.dmg       # Cask
op run --env-file=.env.template -- ./scripts/release.sh 0.1.0 build/mytool          # Formula
```

After release: `brew install <tap>/<name>`.

## Step 4 — subsequent versions

Same wrapping + `release.sh` invocation with the new version.

For ad-hoc tap maintenance:

```bash
jwa-tobrew bump <name>            # syncs to latest GitHub release
jwa-tobrew bump <name> 0.4.2      # syncs to a specific version
```

## Snapshotting an existing release (no source changes)

```bash
jwa-tobrew add github.com/SOMEONE/some-cli
jwa-tobrew add github.com/SOMEONE/some-cli --version 1.2.3 --asset somecli-1.2.3-darwin-arm64.tar.gz
```

`add` is **token-free** — push to the tap uses plain SSH.

## Conventions

- **Conventional commits** (`feat:`, `fix:`, `docs:`, `chore:`) — GoReleaser groups changelog by these prefixes.
- **Semver tags** with `v` prefix (`v0.1.0`).
- One artifact name per project, matching the repo name.
- For Go projects, expose `--version` so the Formula's `test` block passes (the scaffolded GoReleaser config wires `main.version` via ldflags).

## Token handling

If `$GITHUB_TOKEN` is missing, `jwa-tobrew` errors with `"$GITHUB_TOKEN not set — wrap with: jwa-harden run -- jwa-tobrew <command>"`. Solutions in order of preference:

1. `jwa-harden run -- jwa-tobrew <command>` (canonical, once jwa-harden exists)
2. `op run --env-file=.env.template -- jwa-tobrew <command>` (current bootstrap option — the tap repo ships an `.env.template`)
3. Last resort: a transient shell with the env set via `op read`. **Never** export the token to a long-lived shell.

Token rotation is `jwa-harden`'s job, not `jwa-tobrew`'s.

## Troubleshooting

- **"could not locate homebrew-tap clone"** — set `BREWTAP_DIR=/path/to/homebrew-tap`.
- **"tap origin is HTTPS but jwa-tobrew pushes over SSH only"** — `git -C <tap> remote set-url origin git@github.com:OWNER/REPO.git`.
- **goreleaser fails with "git is in a dirty state"** — commit or stash before releasing.
- **Cask install fails with sha256 mismatch** — re-run `jwa-tobrew bump <name>`; the asset was likely re-uploaded.
