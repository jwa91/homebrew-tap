# homebrew-tap

Personal [Homebrew](https://brew.sh) tap for casks and formulae, plus
**`jwa-tobrew`** — the CLI that publishes a project here in one command.

```bash
brew tap jwa91/tap
```

## Installable items

| Name | Kind | Description |
|------|------|-------------|
| [trnscrb](https://github.com/jwa91/trnscrb) | cask | Drop audio, PDFs, and images onto your menu bar to get markdown back |
| jwa-tobrew | formula | Release a project to this tap (see below) |

```bash
brew install --cask jwa91/tap/trnscrb
brew install jwa91/tap/jwa-tobrew
```

## jwa-tobrew

`jwa-tobrew` automates publishing a project to this tap end-to-end: tag, GitHub
Release with auto-generated changelog, asset upload, tap update + commit + push.
It is one of the `jwa-*` family of personal CLIs that share a canonical command
vocabulary (`add`, `bump`, `release`, `init`, `config`, `doctor`, `upgrade`).

**Secrets**: jwa-tobrew never touches 1Password directly. It reads `$GITHUB_TOKEN`
from its environment and errors out if the env is empty when the GitHub API is
needed. Wrap with [`jwa-harden run`](https://github.com/jwa91/homebrew-tap)
(when that exists) or `op run --env-file=.env.template --` for now.

See `~/dotfiles/docs/security-ground-rules.md` for the full security model.

### Commands

```
jwa-tobrew add <github-url>    Snapshot a published GitHub release into the tap (no token)
jwa-tobrew bump <name> [ver]   Re-sync an existing tap entry to a published release
jwa-tobrew release             Tag + release a project + update the tap (run inside the project)
jwa-tobrew init                Scaffold release config in another project
jwa-tobrew config              Regenerate tap.toml + tap.local.toml
jwa-tobrew doctor              Check tools, tap location, SSH origin, env
jwa-tobrew upgrade             Re-install via brew
```

### Adding a new project

In the project repo:

```bash
jwa-tobrew doctor                                        # one-time sanity check
jwa-tobrew init                                          # auto-detects Go / Cask / Formula
op run --env-file=.env.template -- ./scripts/release.sh 0.1.0 [path/to/artifact]
```

For Go projects, GoReleaser handles cross-platform builds, the changelog, and
auto-commits the Formula here. For Casks (macOS apps) and other binaries,
jwa-tobrew handles tag → GitHub release → asset upload → tap update directly.

### Snapshotting an existing release (no source changes needed)

```bash
jwa-tobrew add github.com/SOMEONE/some-cli                   # latest release, auto-detect asset
jwa-tobrew add github.com/SOMEONE/some-cli --version 1.2.3
jwa-tobrew add github.com/SOMEONE/some-app --kind cask --asset SomeApp.dmg
```

This requires no token at all — push to the tap is plain SSH.

## tap.toml + tap.local.toml

The `.rb` files in `Casks/` and `Formula/` are the source of truth. `jwa-tobrew config`
regenerates two computed indexes from them:

- **`tap.toml`** (committed) — manifest of every item in the tap with name, kind,
  source repo, version, and asset pattern. Useful for `jq`-style introspection.
- **`tap.local.toml`** (gitignored) — annotates each item with whether the source
  repo is locally cloned (and where) on this machine.

## Bootstrapping jwa-tobrew (one-time)

Until the first GitHub release of jwa-tobrew exists, `brew install jwa91/tap/jwa-tobrew`
won't work. Bootstrap from a local clone of this repo:

```bash
make install                     # builds and copies to ~/.local/bin/jwa-tobrew
make bootstrap VERSION=0.1.0     # tags + runs goreleaser to publish v0.1.0
git pull --rebase                # pull the auto-committed Formula
brew install jwa91/tap/jwa-tobrew  # now works from anywhere
```

After bootstrapping, future jwa-tobrew releases use jwa-tobrew itself:

```bash
op run --env-file=.env.template -- ./bin/jwa-tobrew release --kind go --version 0.2.0
```

## Layout

```
Casks/                     Homebrew Casks (macOS apps)
Formula/                   Homebrew Formulae (binaries)
cmd/jwa-tobrew/            jwa-tobrew CLI source
  templates/               embedded templates for `jwa-tobrew init`
.env.template              op:// references for GITHUB_TOKEN
.goreleaser.yaml           publishes jwa-tobrew back into this same tap
tap.toml                   generated manifest (committed)
tap.local.toml             generated annotations (gitignored)
Makefile                   build / install / bootstrap targets
```

## Updating already-installed software

```bash
brew update
brew upgrade --cask trnscrb
brew upgrade jwa91/tap/jwa-tobrew
```
