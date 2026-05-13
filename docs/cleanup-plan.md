# Cleanup plan — open ends after agentskills v0.1.0

> Historical note: this snapshot predates the v0.4.1 Cask migration. Current
> family distribution uses `Casks/*.rb` via GoReleaser `homebrew_casks`; see
> ADR 0008 and ADR 0009 for the active contract.

Snapshot from 2026-05-13. After shipping `agentskills` v0.1.0 (commit
`beb2952` on `jwa91/agentskills`), publishing `Formula/agentskills.rb` here
(commit `4bf01d3`), executing the tap-skill migration (commits `f65616c` +
`ba6b959` here, `155bd47` upstream), and codifying the cross-repo contracts
(ADRs 0001–0007 in tap, ADR 0001 in dotfiles, briefing in agentskills) — this
is what's still loose, what we learned, and what we're deliberately leaving
alone.

---

## Key learnings (durable; worth keeping)

1. **goreleaser OSS doesn't parse prefix-tagged versions** (`cli/v0.1.0`,
   `module/v0.1.0`, …) — the `monorepo` feature that solves it is Pro-only.
   Use plain `vX.Y.Z` for binary tags; filter co-existing schemes (per-skill,
   per-module) via `git.ignore_tags`. The `cli/v*` scheme we initially planned
   for `agentskills` was over-engineering for a problem we didn't have.
2. **The release pattern is one trigger, one path.** Tag `vX.Y.Z` locally →
   push tag → GHA `verify → release → smoke` → goreleaser commits
   `Formula/<name>.rb` to the tap → `brew install jwa91/tap/<name>` works on
   a clean machine. Documented as the canonical pattern in `prehandover`'s
   `.github/workflows/release.yml`. Mirror it for every new CLI.
3. **Plan-vs-CLI drift is real and only surfaces during execution.** The
   migration plan called `agentskills link --harness claude`; reality is
   `--project . --harness anthropic --force`. Future migration plans should be
   draft-verified against the actual CLI before sign-off.
4. **`--force` is the idempotency flag for `agentskills`.** Both `bootstrap`
   and `link` error on existing destinations without it. Document loudly in
   the next CLI we build — same precedent.
5. **prek hooks classifier prefixes drift as upstream tools add new
   runtime keys.** Codex's `[hooks.state]` table was new; `sync.py`'s
   `RUNTIME_PREFIXES` hadn't caught up. Drift-detector classifiers need
   maintenance as their upstream evolves.
6. **One-way reference graphs hold.** Tap ADRs cite dotfiles security rules;
   dotfiles never names a tap path. The boundary contract (ADR 0007) is
   self-enforcing as long as the direction rule is honored — and it has been.
7. **The 280-char SKILL.md description cap is load-bearing.** Each installed
   skill's description compounds across sessions and projects. The 4 migrated
   skills landed at 306/321/413/366 chars from the plan's prescribed text;
   trimming to 228/210/237/263 reclaimed ~600 chars per session per project.
8. **One-paragraph ADRs force clarity.** Two bolded labels (`**Scope.**` /
   `**Why.**`), no inner headings, no tables. The constraint is the point.
   `adr-writer` skill codifies the format.
9. **`op read | gh secret set` is the clean pattern for repo secrets.**
   No value ever lands on disk or in logs. Use this for every future tap
   binary that needs `HOMEBREW_TAP_GITHUB_TOKEN` on its repo.

---

## Open ends — high value, low cost (do first)

- [ ] **Commit `~/dotfiles/docs/security-ground-rules.md` to dotfiles.** It's
      currently **untracked**. ADRs 0001/0005/0007 in this tap, dotfiles
      ADR 0001, and four migrated SKILL.md files all reference it as the
      canonical source. Anyone cloning dotfiles fresh today gets nothing.
      Single `git add` + commit.
- [ ] **Clean up dotfiles pre-existing dirty state.** `M config/cursor/settings.json`
      and `M zsh/.zshenv` have been hanging there through the whole session.
      Commit or discard.
- [ ] **Resolve `homebrew-tap` ↔ `~/Developer` casing.** `make bootstrap-skills`
      output showed `/Users/jw/Developer/homebrew-tap` while the canonical path
      is `/Users/jw/developer/homebrew-tap` (lowercase). On macOS this is
      case-insensitive but case-preserving — confusing in logs. Add a symlink
      hint to dotfiles `setup/` or document.
- [ ] **Archive the migration plan's `Pre-migration gates` section.** All
      checkboxes are now satisfied; the section is forward-looking framing.
      Either mark every box ✓ with the commit/date that satisfied it, or move
      the section to a `## Pre-migration gates (satisfied 2026-05-13)`
      header.

## Open ends — medium (schedule for next session)

- [ ] **Bump GHA actions past Node 20** in `agentskills/.github/workflows/`
      and `prehandover/.github/workflows/`. GitHub will force-switch on
      2026-06-02. Bumps: `actions/checkout`, `actions/setup-go`,
      `golangci/golangci-lint-action`, `actions/dependency-review-action`,
      `goreleaser/goreleaser-action`. SHA-pin per security-ground-rules style.
- [ ] **Add a CI workflow to `homebrew-tap`.** Currently only prek hooks fire
      locally; PRs and main pushes have no GHA verification. Inconsistent
      with the Go-CLI repos. Mirror `agentskills/.github/workflows/ci.yml`
      (single `check` job: setup-go → `make check`). Also add a minimal
      `align` job that runs `go run ./tools/jwa-tobrew align`.
- [ ] **Bump `tools/agentskills/cmd` test coverage** from 10.7% toward 50%+.
      The command dispatch is thin but flag parsing + interactive picker
      paths could regress silently.
- [ ] **First real use of the `adr-writer` skill.** It exists, but ADRs
      0001–0007 + dotfiles 0001 were all written before it existed. Next
      ADR (whatever it is) should go through `adr-writer` to validate the
      skill's body actually scaffolds usable output.
- [ ] **Decide fate of the Python CLI under `agentskills/src/agentskills/`.**
      Briefing said "keep until a skill author hits friction." Has anyone?
      If not, leave; if so, port or drop.
- [ ] **Cut a version of `homebrew-tap`.** `[Unreleased]` has accumulated
      8 entries this session. The tap has its own SemVer (per CHANGELOG
      header — Major/Minor/Patch defined). Cut whichever bump fits.
- [ ] **Per-skill release tags (`<skill>/vX.Y.Z`) are filtered but unused.**
      `git.ignore_tags` in `agentskills/.goreleaser.yaml` is set up to make
      them safe. First skill that needs versioning activates the lane.

## Open ends — watch-only (don't fix preemptively)

- goreleaser's `brews` deprecation warning. Comment in `.goreleaser.yaml`
  documents *why* `brews:` stays (Formula, not Cask). When goreleaser actually
  removes the block, migrate at that point.
- `jwa-harden` doesn't exist yet but is named by ADRs 0001/0007 + multiple
  skills as the canonical secret wrapper. Track; don't build until `op run`
  friction is real.
- The Brewfile reservation for `jwa-tobrew` is still commented out
  (line 15 of dotfiles `Brewfile`). `jwa-tobrew` ships from this tap but
  isn't yet pinned to a tagged release. Activate when `jwa-tobrew` cuts
  its own `vX.Y.Z` via the same pattern.

---

## Out of scope (deliberately not doing — not bugs)

- **Renaming any ADR.** Numbers are append-only. A superseded ADR gets a
  "Superseded by NNNN" line prepended; it stays at its number.
- **Auto-link in `agentskills bootstrap`.** Briefing context-bloat rule #2:
  install is opt-in, no invisible state.
- **Embedding skill content in the Go binary.** ADR 0006 forbids; the
  CLI clones at install time.
- **Adding a `dependencies` field to skills.** Briefing context-bloat
  rule #5: no install-time dependency chains.
- **Vendoring binaries from the tap into dotfiles `Brewfile` as
  alternates.** ADR 0007 forbidden pattern (c).
