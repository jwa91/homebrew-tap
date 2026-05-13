---
name: tap-alignment
description: Detect and fix drift between this repo (or a project that publishes to it) and the conventions encoded in `jwa-tobrew`, prek, and the ADRs. Trigger when the user says "align", "is this in line", "any drift", "verify conventions", "what changed since the last convention update", or asks why a particular file/symlink/script is required. Also use this skill when a `prek` hook fails on the `jwa-tobrew-align` step.
---

# Keeping the tap and its projects aligned

The tap codifies conventions in three layers, and `jwa-tobrew align` is the tool that detects when reality drifts from any of them.

## The three layers

1. **ADRs** at `docs/adr/` — *why* a convention exists. Read these before changing one. Five live as of v0.2.0; numbered, never renumbered.
2. **`align` rules** in `tools/jwa-tobrew/align.go` — *what* the convention requires, with auto-fix when safe. Detects mode (tap vs project) from cwd contents; emits findings, optionally applies them.
3. **`prek.toml` hooks** — *when* the rules run. The `jwa-tobrew-align` hook re-runs `align` on every commit that touches paths that affect tap conventions. The conventional-commit hook runs on every commit message.

A new convention should be added to all three: ADR (why) → align rule (what) → prek-hook scope adjustment (when).

## The two modes

`align` looks at the cwd and picks one:

- **tap mode** — when both `Casks/` and `Formula/` exist. Checks: `.agents/skills/jwa-tobrew/` exists with `.claude/skills/jwa-tobrew` as a symlink to it; `docs/adr/` exists; README has `<!-- BEGIN ITEMS -->` / `<!-- END ITEMS -->` markers; `CHANGELOG.md` exists; if the tap embeds a Go CLI (`go.mod` at root), `.goreleaser.yaml` is also at root.
- **project mode** — when `go.mod`, `Package.swift`, or an `*.xcodeproj`/`*.xcworkspace` exists. Checks: `.env.template` is present; `.gitignore` blocks `.env*`; `.agents/skills/release/` exists with a `.claude` symlink (when present); per kind: Go → `.goreleaser.yaml`; cask → `scripts/release.sh`.

If neither set of markers is found, `align` errors with "not a tap and not a recognised project".

## Findings

Each finding has one of two shapes:

- **Auto-applyable** — `align` can fix it (file creation, symlink creation, `.gitignore` append). Reports as `→ <path> — <action>`. Apply with `--apply`.
- **Manual** — needs flag input the command can't infer (e.g. `init --kind=cask` to scaffold). Reports as `! <path> — ... [manual]`. Run the suggested command yourself.

`align` exits non-zero if any finding remains after the run; that's what makes it suitable as a CI check or pre-commit hook.

## When to run align

- After touching anything under `Casks/`, `Formula/`, `.agents/skills/`, `docs/adr/`, `tools/jwa-tobrew/`, `README.md`, or `CHANGELOG.md`. (`prek` does this automatically when the hook is installed.)
- After `jwa-tobrew init` in a fresh project — confirms the scaffold is complete.
- Before opening a PR or committing structural changes.
- As the first triage step when "something feels off" with the repo layout.

## What is *not* enforced today (deliberately)

- **`.env.template` content** — the file's existence is checked but the `op://` reference's value isn't. If the user moves to a different secret-handling model later (e.g. `jwa-harden run`), the env-template requirement may be dropped or replaced — don't promote it as eternal.
- **`.rb` schema** — desc, homepage, semver version, github source, filename↔name match. Queued for `doctor --strict` (TODO #2 in `TODO.md`).
- **`tap.toml` JSON-schema validation** — Queued for TODO #4.
- **Sha verification of every committed `.rb` against its GitHub release** — Queued for TODO #1, will run as a daily CI cron.

When you read a finding that feels wrong, check this list — it might be a known gap rather than a real drift.

## Adding a new convention

1. Decide *why* — write a one-paragraph ADR at `docs/adr/000N-<slug>.md`. If you don't have a "why", the convention is probably not worth adding.
2. Encode it in `tools/jwa-tobrew/align.go` (in `alignTap` or `alignProject`). Auto-applyable if reasonable; manual otherwise.
3. Adjust the `jwa-tobrew-align` hook's `files:` regex in `prek.toml` so the hook fires when the relevant paths change.
4. Run `prek run --all-files` to verify the existing repo doesn't drift; fix or `--apply`.
5. Add a CHANGELOG entry under `Unreleased`.
6. Update relevant skills (often this one, plus `scaffold-cli` if the convention affects new projects).

## Common drift patterns and what they mean

- **`.claude/skills/<name>` is a real dir, not a symlink** — happens when a skill is created via Claude tooling and not migrated. Fix: `align --apply` migrates it to `.agents/skills/<name>/` and symlinks back.
- **README has stale items table** — happens when an `.rb` is hand-edited without running `jwa-tobrew config`. Fix: `jwa-tobrew config` regenerates and you commit the diff.
- **`tap.toml` and `.rb`s disagree** — same cause as above. `config` is the one-shot.
- **Multi-asset `.rb` flagged in tap mode** — not flagged by align (fine to exist), but `bump`/`release` will refuse it. The source repo's GoReleaser owns updates; that's intentional, not drift.
