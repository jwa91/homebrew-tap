# Migration plan — move tap-related skills into `agentskills`

> Historical note: this plan was written during the Formula-to-Cask transition.
> The current contract is ADR 0006 + ADR 0009, and `agentskills` now publishes
> `Casks/agentskills.rb` via GoReleaser `homebrew_casks`.

This plan moves the four tap-related skills currently vendored under
`/Users/jw/developer/homebrew-tap/.agents/skills/` into the `agentskills`
monorepo at `/Users/jw/developer/ai-monorepo/agentskills/skills/`, and converts
the tap's `.agents/skills/` tree into a generated, gitignored artifact
materialised by `agentskills bootstrap` on a fresh clone.

Anchor: ADR 0006 (`docs/adr/0006-agentskills-boundary-contract.md`). One
direction at the file level (agentskills' goreleaser writes `Formula/agentskills.rb`
into this tap), one direction at the install level (this tap's bootstrap invokes
`agentskills bootstrap …`). Skill content is **never** embedded in the Go
binary — it is fetched at install time from `jwa91/agentskills`.

**Status: executed 2026-05-13** in homebrew-tap commit `f65616c` and agentskills
commit `155bd47`. The snippets and notes below were reconciled against the
shipped `Makefile` / `.gitignore` after execution — they now describe what was
actually done. If you find a discrepancy between this doc and the committed
files, the committed files are the source of truth.

---

## Pre-migration gates

Every box must be ticked before step 1 (file moves) runs. If any fails, stop
and fix it before continuing — partial migration leaves the tap unable to
bootstrap its own align skill.

- [ ] **`vX.Y.Z` tag exists on `jwa91/agentskills`.** `gh release view vX.Y.Z --repo jwa91/agentskills` returns a release with goreleaser-built darwin+linux × amd64+arm64 archives and a `checksums.txt`. Tag scheme is plain `v*` for the binary; per-skill tags (`<skill>/vX.Y.Z`) are filtered by goreleaser's `git.ignore_tags`.
- [ ] **`Formula/agentskills.rb` is present in this tap on `main`** and was written by `agentskills`' goreleaser `brews:` block (not by hand). Filename matches the binary name (`agentskills`). Source = `https://github.com/jwa91/agentskills` at the `vX.Y.Z` tag. Multi-platform (4 `url` lines) so `jwa-tobrew bump`/`release` correctly refuse it as source-repo-managed (per `jwa-tobrew` SKILL.md §Two release patterns).
- [ ] **`brew install jwa91/tap/agentskills` works on a clean machine.** Verified in either (a) a fresh macOS VM, (b) a Linuxbrew container, or (c) a clean shell with `HOMEBREW_NO_INSTALL_FROM_API=1`, `brew untap jwa91/tap || true`, `brew uninstall agentskills || true`, then `brew tap jwa91/tap && brew install agentskills`. `agentskills --version` returns the expected `vX.Y.Z`.
- [ ] **`agentskills doctor` is green** on the same machine: git on PATH, cache dir writable at `~/.cache/agentskills/repos/`, repo `https://github.com/jwa91/agentskills.git` reachable, default ref `main` resolvable.
- [ ] **`agentskills bootstrap --skill jwa-tobrew` succeeds against the agentskills repo HEAD** (post-step-7 of the briefing migration order). Run it in a scratch dir: `mkdir /tmp/asbt && cd /tmp/asbt && agentskills bootstrap --skill jwa-tobrew`. Result: `/tmp/asbt/.agents/skills/jwa-tobrew/SKILL.md` exists with frontmatter `name: jwa-tobrew`. Repeat for `scaffold-cli`, `tap-alignment`, `tap-scaffolding`.
- [ ] **All four skills exist under `agentskills/skills/<name>/`** in the agentskills repo at the ref the tap will pin (`main`, or an explicit `--ref` we choose to set in the bootstrap target). Verified with `agentskills list` against that ref.
- [ ] **The agentskills repo has per-skill release tags** for at least `jwa-tobrew/vX.Y.Z` and `tap-alignment/vX.Y.Z` (skills can release on their own cadence per ADR 0006). Not strictly required for bootstrap (which can pin a branch), but required for the rollback plan to work cleanly.
- [ ] **A holdout copy of the current `.agents/skills/` tree is preserved on a branch** (`pre-agentskills-migration`) in this repo, so rollback (§Rollback) is one `git checkout` away. Don't rely solely on history — keep the branch until the migration has lived on `main` for a release cycle.

---

## File moves

Source repo: `/Users/jw/developer/homebrew-tap/`.
Destination repo: `/Users/jw/developer/ai-monorepo/agentskills/`.

For each of the four skills the move is the same shape. Copy the directory
verbatim (SKILL.md plus any future scripts/references/assets the skill grows),
keep the directory name identical (CLI requires `frontmatter.name == dirname`,
briefing §Skill structure conventions), then delete the source dir and the
matching `.claude/skills/<name>` symlink from this tap.

| Skill | Source dir (tap) | Destination dir (agentskills) | Symlink to remove from tap |
|---|---|---|---|
| `jwa-tobrew` | `/Users/jw/developer/homebrew-tap/.agents/skills/jwa-tobrew/` | `/Users/jw/developer/ai-monorepo/agentskills/skills/jwa-tobrew/` | `/Users/jw/developer/homebrew-tap/.claude/skills/jwa-tobrew` |
| `scaffold-cli` | `/Users/jw/developer/homebrew-tap/.agents/skills/scaffold-cli/` | `/Users/jw/developer/ai-monorepo/agentskills/skills/scaffold-cli/` | `/Users/jw/developer/homebrew-tap/.claude/skills/scaffold-cli` |
| `tap-alignment` | `/Users/jw/developer/homebrew-tap/.agents/skills/tap-alignment/` | `/Users/jw/developer/ai-monorepo/agentskills/skills/tap-alignment/` | `/Users/jw/developer/homebrew-tap/.claude/skills/tap-alignment` |
| `tap-scaffolding` | `/Users/jw/developer/homebrew-tap/.agents/skills/tap-scaffolding/` | `/Users/jw/developer/ai-monorepo/agentskills/skills/tap-scaffolding/` | `/Users/jw/developer/homebrew-tap/.claude/skills/tap-scaffolding` |

**Copied verbatim** (one `cp -R` per skill into `agentskills/skills/`):

- `SKILL.md` (with the description-only edits in the next section — applied
  *in the destination repo*, not in the source).
- Any sibling files/dirs that currently exist or get added later: `scripts/`,
  `references/`, `assets/`. As of today all four skills are SKILL.md-only;
  the contract still holds.

**Recreated, not copied:**

- Nothing. The four skills have no generated artifacts; they are pure content.
  In particular: the `.claude/skills/<name>` symlinks are **not** moved — the
  agentskills repo doesn't ship harness adapters in-tree (that's
  `internal/harness/harness_adapters.json` inside the CLI). Re-creation of
  harness symlinks at the consumer site is `agentskills link`'s job, run
  separately after bootstrap.

**Removed from the tap (in the same commit that adds the bootstrap target):**

1. `rm -rf .agents/skills/jwa-tobrew .agents/skills/scaffold-cli .agents/skills/tap-alignment .agents/skills/tap-scaffolding`
2. `rm .claude/skills/jwa-tobrew .claude/skills/scaffold-cli .claude/skills/tap-alignment .claude/skills/tap-scaffolding` (these are symlinks; `rm`, not `rm -rf`)
3. If `.agents/skills/` and `.claude/skills/` end up empty, keep the parent
   dirs — the bootstrap target re-populates them and git won't track empty
   dirs anyway.

Order: **destination first** (commit lands on `agentskills` `main`, per-skill
`<name>/vX.Y.Z` tags cut), **then** source removal in this tap. That way the
tap's bootstrap target works on the migration commit itself.

---

## SKILL.md edits during the move

These are description-frontmatter-only edits, applied in the destination repo
as part of the move. Bodies are out of scope. Rule references are to the
"Context-bloat discipline" section of `agentskills/docs/cli-go-port-briefing.md`.

### `skills/jwa-tobrew/SKILL.md`

- **Current description** names three other skills (`scaffold-cli`,
  `tap-alignment`, and indirectly the `jwa-harden`/`jwa-vps` siblings that
  appear in the body, not the description — those are fine).
- **Rule #4 violation:** "a skill may name at most one related skill in its
  description." Today: "For new project setup walk-throughs, also check the
  `scaffold-cli` skill; for drift detection, see `tap-alignment`." → trim to
  **one**.
- **Edit:** drop the `scaffold-cli` reference; keep `tap-alignment`, since
  drift detection is the cross-skill link an agent is most likely to need
  while another skill is active (alignment runs as a `prek` hook). Move the
  `scaffold-cli` pointer into the SKILL.md body under a `## Related` heading
  (body edits are scoped to the migration commit but covered separately —
  this section is description-only).
- **Char-cap check:** trimmed description still needs to be ≤ ~280 chars
  (rule #1).

Proposed new description (single line, ~270 chars, trigger-shaped, names
exactly one related skill):

> Use the `jwa-tobrew` CLI to publish a project to the user's personal Homebrew tap (jwa91/tap). Trigger when the user says "publish to my tap", "add to brew", "make this installable via brew", "release this to my tap", or asks about the `jwa-tobrew` command surface. For drift detection see `tap-alignment`.

### `skills/scaffold-cli/SKILL.md`

- **Current description** names two related skills: `jwa-tobrew` ("For the
  existing CLI surface see `jwa-tobrew`") and `tap-alignment` ("for convention
  enforcement see `tap-alignment`"). Two = rule #4 violation.
- **Edit:** keep `jwa-tobrew` (it's the CLI this skill is teaching the user
  to install and use); drop `tap-alignment` from the description. Move the
  alignment pointer into the body.
- **Char-cap check:** result is still under ~280 chars.

Proposed new description:

> Set up a new Go CLI to be installable via the user's personal Homebrew tap (jwa91/tap). Trigger when the user says "new CLI", "scaffold a CLI", "start a Go project for the tap", "make a new tool installable via brew", or asks how to bootstrap a publishable CLI from scratch. For the existing CLI surface see `jwa-tobrew`.

### `skills/tap-alignment/SKILL.md`

- **Current description** names zero related skills. Already compliant with
  rule #4 (zero ≤ one). Char count is well under cap. **No edit needed.**

### `skills/tap-scaffolding/SKILL.md`

- **Current description** names two: `scaffold-cli` and `jwa-tobrew`
  ("end-users wanting to scaffold a project should use `scaffold-cli` (Go)
  or `jwa-tobrew init --kind=cask` (cask)"). Two = rule #4 violation.
- This skill is for *maintainers of jwa-tobrew* — the disclaimer is
  load-bearing because the audience is narrow. Keep the audience disclaimer
  but collapse to one pointer.
- **Edit:** keep `scaffold-cli` (most common wrong-skill mistake for an
  end-user); drop the `jwa-tobrew init --kind=cask` parenthetical (that's a
  command surface, not a skill reference — technically allowed but the
  intent of rule #4 is to keep the list short, and the body already covers
  cask scaffolding).

Proposed new description:

> Modify or extend the scaffolding system inside `jwa-tobrew` (the templates that `init` writes into target projects). Trigger when the user says "add a new scaffold kind", "change what init writes", "update the templates", "add a new file to scaffolds", or asks how the embedded templates are wired. This skill is for maintainers; end-users should use `scaffold-cli`.

### Summary

| Skill | Current refs | Trimmed to | Rule #4 fix |
|---|---|---|---|
| `jwa-tobrew` | `scaffold-cli`, `tap-alignment` | `tap-alignment` | drop `scaffold-cli` |
| `scaffold-cli` | `jwa-tobrew`, `tap-alignment` | `jwa-tobrew` | drop `tap-alignment` |
| `tap-alignment` | (none) | (none) | already compliant |
| `tap-scaffolding` | `scaffold-cli`, `jwa-tobrew` | `scaffold-cli` | drop `jwa-tobrew` |

---

## align.go changes (planned, not executed)

Reference: `/Users/jw/developer/homebrew-tap/tools/jwa-tobrew/align.go`.

Today `align` enforces in `alignTap` that `.agents/skills/jwa-tobrew/` exists
locally with a `.claude/skills/jwa-tobrew` symlink pointing at it. After
migration, those dirs become generated-and-gitignored artifacts produced by
`agentskills bootstrap`. The drift rule shifts from "this directory exists
in the working tree" to "this tap's bootstrap mechanism can produce it on
demand."

Function-by-function:

### `alignTap(dir string)` — `align.go:156`

- **Line 158: `f = append(f, skillSymlinkFindings(dir, "jwa-tobrew")...)`** —
  this is the only call site that names a tap-managed skill explicitly.
  - **Change to:** remove this line, replace with a single new finding
    `bootstrapSkillsAvailable(dir)` that checks the bootstrap mechanism
    (next subsection), not the materialised dir.
  - **Rationale:** ADR 0006 — drift is no longer "do these dirs match across
    repos?" but "does `agentskills bootstrap` succeed against the pinned
    ref?". The dirs themselves are generated.

### New finding (planned, to be added in `alignTap`)

Add a new finding that:

1. Checks `Makefile` (or `bin/`) for the `bootstrap-skills` target (see
   §Tap repo bootstrap entry) — string-grep for the target name.
2. Checks that `agentskills` is on `PATH` (`exec.LookPath("agentskills")`).
   If absent, finding action: "install via `brew install jwa91/tap/agentskills`".
3. Auto-fix: `make bootstrap-skills` (or invoke the script directly). The
   target is idempotent (§Tap repo bootstrap entry), so re-running on an
   already-bootstrapped tree is a no-op.

Finding ID (loose convention — `align.go` doesn't use stable IDs today, just
`Path` + `Action` strings): `Path: ".agents/skills/"`, `Action: "missing
tap-managed skills; run \`make bootstrap-skills\`"`. Use the same finding
shape as the existing entries (`alignFinding` struct).

### `skillSymlinkFindings(root, name string)` — `align.go:101`

- **Keep as-is.** It's still called from `alignProject` at `align.go:220`
  for the project-level `release` skill, which is *project-managed*, not
  tap-managed, and is the skill `jwa-tobrew init` writes via `writeAuxiliary`
  (per `tap-scaffolding/SKILL.md`). That skill stays vendored per-project;
  it's not part of this migration.
- The function itself is general-purpose (`name` is a parameter); nothing
  about its body changes.

### `alignProject(dir string)` — `align.go:198`

- **No change.** The only `skillSymlinkFindings` call in project mode
  (`align.go:220`) is for `release`, which is out of scope here.

### `detectAlignMode`, `runAlign` — `align.go:81`, `align.go:20`

- **No change.** Mode detection is content-based (`Casks/ + Formula/`) and
  doesn't reference any specific skill.

### Summary table

| Location | Rule today | Decision | Detail |
|---|---|---|---|
| `align.go:158` | `skillSymlinkFindings(dir, "jwa-tobrew")` inside `alignTap` | **Remove** | Replaced by bootstrap-availability finding. |
| `align.go` (new, inside `alignTap`) | (none) | **Add** | `bootstrap-skills` Make target present + `agentskills` on PATH; auto-fix runs the target. |
| `align.go:220` | `skillSymlinkFindings(dir, "release")` inside `alignProject` | **Keep** | `release` is project-managed, not tap-managed. Out of scope. |
| `align.go:101` | `skillSymlinkFindings` helper | **Keep** | Still used by `alignProject`. |
| `align.go:81` | `detectAlignMode` | **Keep** | Content-based, skill-agnostic. |
| `prek.toml:68` regex | Triggers `jwa-tobrew-align` on `^(\.agents/skills/|\.claude/skills/|...)` changes | **Keep** | Regenerated skills tree is gitignored (next section), so it won't appear in staged paths anyway; regex stays harmless. |

---

## Tap repo bootstrap entry

Add to `/Users/jw/developer/homebrew-tap/Makefile`. Idempotent: re-running
on an already-populated tree does not error; `agentskills bootstrap`'s default
mode (`copy`) overwrites the target dir contents but does not delete sibling
files, and the harness symlinks (`agentskills link`) check existence before
creating.

Exact target (to be added; the Makefile already has `.PHONY: build install
dev bootstrap clean hooks` — extend it):

```make
# Pinned ref the tap installs from. Policy (see Decisions): commit a SHA
# from jwa91/agentskills's main branch. Bumping the pin is an explicit,
# reviewable commit in this tap. `main` is permitted via env override
# (`AGENTSKILLS_REF=main make bootstrap-skills`) for local development,
# but the committed default must never be `main` — that would re-create
# the dual-source-of-truth drift ADR 0006 exists to prevent.
AGENTSKILLS_REF ?= 155bd475f9c608c2f9240f0dfb20f3c102211f1a

# Populate .agents/skills/ from the agentskills repo. Idempotent — safe to
# re-run on a fresh clone or after pulling. .agents/skills/ and
# .claude/skills are gitignored; this target is how they get created.
bootstrap-skills:
	@which agentskills >/dev/null 2>&1 || (echo "install: brew install jwa91/tap/agentskills" && exit 1)
	agentskills bootstrap \
	  --skill jwa-tobrew \
	  --skill scaffold-cli \
	  --skill tap-alignment \
	  --skill tap-scaffolding \
	  --ref $(AGENTSKILLS_REF) \
	  --mode copy \
	  --force
	agentskills link --project . --harness anthropic --force
	@echo "tap-managed skills materialised under .agents/skills/ and linked into .claude/skills/"
```

Notes:

- Also extend the existing `.PHONY:` line: add `bootstrap-skills`.
- Add `bootstrap-skills` as a dependency of a new `setup` aggregate target
  (`setup: hooks bootstrap-skills`) so a fresh clone is `make setup` away
  from a working tree. Don't make it a dependency of `build` — Go builds
  must not require network access.
- `--mode copy` is the default per the briefing; explicit for documentation.
- `--force` on both `bootstrap` and `link` is required for idempotency:
  without it, re-running the target on an already-populated tree errors
  ("destination already exists"). With `--force`, second runs are clean
  no-ops.
- `--ref $(AGENTSKILLS_REF)` pins a SHA from `jwa91/agentskills`'s
  `main` branch (see Decisions for rationale). The agentskills CLI's
  own default of `main` does **not** apply here: an unpinned default
  in a committed Makefile would let an upstream push silently change
  what a fresh tap clone bootstraps.
- `agentskills link --project . --harness anthropic` creates a **single**
  symlink `.claude/skills → ../.agents/skills`, not per-skill symlinks
  (the legacy shape was per-skill; the current CLI uses one symlink to
  the parent). The harness key is `anthropic` (not `claude`) — adapter
  table is in `internal/harness/harness_adapters.json`.

---

## .gitignore additions

Append to `/Users/jw/developer/homebrew-tap/.gitignore` (after the existing
"Real .env files are forbidden" block). The whole `.agents/skills/` tree
and the `.claude/skills` symlink become generated artifacts of
`make bootstrap-skills`.

Exact lines to add:

```gitignore

# Tap-managed skills are materialised by `make bootstrap-skills`, which
# invokes `agentskills bootstrap` against jwa91/agentskills. The skill
# content lives there; here it's a generated artifact (ADR 0006).
# .claude/skills may exist as a directory (legacy) or as a single
# symlink to ../.agents/skills (current `agentskills link` shape) — the
# no-trailing-slash form covers both.
/.agents/skills/
/.claude/skills
```

Notes:

- Leading slash anchors to repo root, so a future `tools/foo/.agents/skills/`
  (unlikely, but possible) is not affected.
- We gitignore the whole `.agents/skills/` and `.claude/skills/` trees, not
  the parents. The parents (`.agents/`, `.claude/`) keep any non-skills
  content if it's added later.
- Project-managed skills under `tools/<name>/.agents/skills/release/` are
  unaffected — the patterns are repo-root-anchored.

If the migration ever needs to ship a non-tap-managed file under
`.agents/skills/` (none planned), switch to an allow-list pattern with `!`
entries.

---

## Rollback

If `agentskills` proves unreliable in production (CLI bugs, repo
unreachability, ref-pinning issues) and the tap needs its skills back as
vendored content:

1. **Identify the migration commit** on `main` of this tap. Call it
   `MIGRATION_SHA`. It is the commit that:
   - removed the four skill dirs from `.agents/skills/`,
   - removed the four `.claude/skills/<name>` symlinks,
   - added `bootstrap-skills` to the `Makefile`,
   - added `.agents/skills/` and `.claude/skills/` to `.gitignore`,
   - modified `tools/jwa-tobrew/align.go` per §align.go changes.
2. **Revert it:** `git revert --no-commit MIGRATION_SHA && git commit -m
   "revert: roll back agentskills migration"`. This restores the four skill
   dirs from history (they're in tree pre-migration), restores their
   symlinks, restores the `align.go` rules, and removes the `.gitignore`
   entries and the Make target.
3. **If the agentskills repo has diverged** (the skills there got real
   edits after the move and the tap needs them): cherry-pick those updates
   manually, file-by-file, from `agentskills/skills/<name>/SKILL.md` into
   `homebrew-tap/.agents/skills/<name>/SKILL.md`. There is no automated
   re-vendoring path — accept manual reconciliation as the cost of rollback.
4. **Re-create the symlinks if `git revert` doesn't restore them cleanly**
   (symlinks-in-history can be finicky across machines): from the tap root,
   for each `<name>` in `{jwa-tobrew, scaffold-cli, tap-alignment, tap-scaffolding}`:
   `ln -s ../../.agents/skills/<name> .claude/skills/<name>`. Then
   `./bin/jwa-tobrew align` to confirm the four `skillSymlinkFindings`
   reports return clean.
5. **Disable the new bootstrap target** by deleting (not just commenting)
   the `bootstrap-skills` target and `AGENTSKILLS_REF` var from the
   `Makefile`. Leaving it half-defined is worse than removing it — users
   may invoke it on a tree where it now corrupts the vendored content.
6. **Keep `Formula/agentskills.rb`** in place regardless. The boundary
   contract (ADR 0006) is unidirectional at the file level; the tap is
   still a fine distribution channel for the CLI even if we choose not to
   consume it locally. Removing the formula would break other users of the
   tap.
7. **Holdout branch:** the `pre-agentskills-migration` branch from
   §Pre-migration gates is the safety net if `git revert` itself fails.
   `git checkout pre-agentskills-migration -- .agents/skills/ .claude/skills/`
   to restore the trees verbatim, then commit.

---

## Decisions

- **Pinning policy for `AGENTSKILLS_REF`: SHA-pin in the `Makefile`
  variable.** The committed default is a specific commit SHA from
  `jwa91/agentskills`'s `main` branch; bumping it is an explicit,
  reviewable commit in this tap. Per-skill tags
  (`jwa-tobrew/v0.3.1`, …) cannot compose into a single `--ref` call
  — `agentskills bootstrap` takes one ref and applies it to the whole
  repo, so four skills would need four invocations (or a synthesised
  bootstrap tag stream). Inventing a tap-bootstrap tag stream over
  the same content re-creates the dual-source-of-truth drift that
  ADR 0006 was written to kill. `main` is fine for local dev via env
  override (`AGENTSKILLS_REF=main make bootstrap-skills`) but the
  committed form must not default to `main`, because an upstream push
  would otherwise silently change what `make bootstrap-skills`
  materialises on a fresh tap clone. The actual SHA gets set after
  `v0.1.0` lands and the four skills exist on `agentskills` `main`.

## Open questions

_(None currently.)_
