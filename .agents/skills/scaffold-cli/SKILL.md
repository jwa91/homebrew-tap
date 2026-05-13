---
name: scaffold-cli
description: Set up a new Go CLI to be installable via the user's personal Homebrew tap (jwa91/tap). Trigger when the user says "new CLI", "scaffold a CLI", "start a Go project for the tap", "make a new tool installable via brew", "I want to publish a new command", or asks how to bootstrap a publishable CLI from scratch. Walks through repo creation, scaffolding, first release, and follow-on releases. For the existing CLI surface see `jwa-tobrew`; for convention enforcement see `tap-alignment`.
---

# Scaffold a new Go CLI publishable via the tap

Goal: end up with a Go CLI whose `goreleaser release` builds cross-platform binaries, creates a GitHub Release with notes, and writes `Formula/<name>.rb` directly to <https://github.com/jwa91/homebrew-tap>. After the first release, `brew install jwa91/tap/<name>` works from any machine.

## Preconditions

- `gh`, `git`, `goreleaser`, `op` (1Password CLI), and `jwa-tobrew` on PATH (`jwa-tobrew doctor` will check)
- A clone of `homebrew-tap` somewhere reachable (`$BREWTAP_DIR` or one of the standard locations: `~/developer/homebrew-tap`, `~/Developer/homebrew-tap`, `~/src/homebrew-tap`, `~/code/homebrew-tap`)
- The tap clone's `origin` is SSH (not HTTPS) — see ADR 0002
- `$GITHUB_TOKEN` available via 1Password (the tap repo's `.env.template` shows the canonical reference)

## Step 1 — create the source repo

Standard `gh repo create` flow; nothing CLI-specific yet:

```bash
mkdir <name> && cd <name>
git init
go mod init github.com/<owner>/<name>
echo 'package main\n\nimport "fmt"\n\nfunc main() { fmt.Println("hello") }' > main.go
git add . && git commit -m "feat: initial commit"
gh repo create <name> --public --source=. --remote=origin --push
```

`name` should match the binary you want users to type (so `brew install jwa91/tap/<name>` works).

## Step 2 — scaffold the release config

```bash
jwa-tobrew init
```

`init` auto-detects Go (because of `go.mod`) and writes:

- **`.goreleaser.yaml`** — the entire release pipeline. Builds darwin/linux × amd64/arm64, creates the GitHub Release, writes `Formula/<name>.rb` to the tap via the `brews:` block. Same config runs locally and in CI.
- **`.env.template`** — `op://` references for `$GITHUB_TOKEN` (current bridge until `jwa-harden` lands; see security note in the `jwa-tobrew` skill).
- **`.gitignore`** entries that block raw `.env` files (per ADR 0005).
- **`.agents/skills/release/SKILL.md`** + `.claude/skills/release` symlink — a project-level release skill so future agents know how this project ships.

Useful flags: `--name <name>` (defaults to repo name), `--desc "..."` (used in the Formula's `desc` line), `--force` (overwrite existing files).

## Step 3 — wire the Go CLI for `--version`

The scaffolded `.goreleaser.yaml` injects build info via ldflags into `main.version`, `main.commit`, `main.date`. Mirror these in your `main.go`:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
        fmt.Printf("%s %s (commit %s, built %s)\n", os.Args[0], version, commit, date)
        return
    }
    // ... rest of main
}
```

The Formula's auto-generated `test` block runs `<name> --version` and asserts the version string matches; without these vars the test fails on `brew install`.

## Step 4 — first release (local)

```bash
git tag -a v0.1.0 -m v0.1.0
git push origin v0.1.0
op run --env-file=.env.template -- goreleaser release --clean
```

GoReleaser will:

1. Verify HEAD's tag matches `v0.1.0`
2. Build the cross-platform binaries
3. Create the `v0.1.0` GitHub Release with auto-grouped changelog (Conventional Commits → "Features" / "Bug fixes" / "Other")
4. Upload archives + `checksums.txt`
5. Author a commit on `homebrew-tap/main` adding `Formula/<name>.rb` with all four `url`+`sha256` pairs

After it lands, `cd ~/developer/homebrew-tap && git pull` brings the new Formula in. The README items table updates next time anyone runs `jwa-tobrew config` (or auto-updates on the next `add`/`bump`/`release`).

## Step 5 — first release (CI alternative)

For tag-driven CI release, add a workflow at `.github/workflows/release.yml`:

```yaml
name: release
on:
  push:
    tags: ['v*.*.*']
permissions:
  contents: write
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version: stable }
      - uses: goreleaser/goreleaser-action@v6
        with: { args: release --clean }
        env:
          GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

`HOMEBREW_TAP_TOKEN` is a fine-grained PAT with **write** access to `jwa91/homebrew-tap` and **contents:write** on this repo. Local releases use `op run`, CI uses repo secrets — same `goreleaser release` invocation either way.

## Step 6 — subsequent releases

Same as step 4 with a new tag, or just push a new `v*.*.*` tag if CI is wired.

```bash
git tag -a v0.2.0 -m v0.2.0
git push origin v0.2.0
# (CI handles it from here)
```

## What you get

- Installable on any Mac/Linux: `brew install jwa91/tap/<name>`
- Updates flow when you `git push origin v0.X.Y` — no manual tap edit
- The CLI participates in this tap's conventions automatically (the scaffold leaves it in a state `jwa-tobrew align` considers clean)

## What this skill is NOT

- **Not for casks or non-Go formulas** — those don't use GoReleaser; their release flow goes through `scripts/release.sh` (which wraps `jwa-tobrew release`). Use `jwa-tobrew init --kind=cask` (or `--kind=formula`) and follow the hint it prints.
- **Not for snapshotting an existing release into the tap** — for that, `jwa-tobrew add github.com/SOMEONE/some-cli` is the one-shot, no source-repo changes needed.
- **Not the place for token storage details** — that's `~/dotfiles/docs/security-ground-rules.md`. The transitional `.env.template` + `op run` pattern is documented there too.
