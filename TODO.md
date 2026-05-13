# TODO — homebrew-tap

Project-scoped tasks. Cross-repo work lives in `~/dotfiles/TODO.md`.

---

## Queued

- [ ] **Tap repo hardening** — make sure only I can write to the tap and that
      every committed `.rb` traces back to a real, verified GitHub release.
      Concretely:
      - Branch protection on `main`: require PR, require signed commits,
        restrict force-push, restrict who can merge to me only.
      - `CODEOWNERS` claiming everything in `Casks/`, `Formula/`, `cmd/`.
      - `gitleaks` config + a CI job that runs it on every PR.
      - A CI job (`scripts/verify-tap.sh`?) that for each `.rb`:
          (a) confirms a GitHub release at the pinned `version` tag exists,
          (b) re-fetches the asset and compares sha256,
          (c) fails the PR if either check fails.
      - A CI job that fails if any tracked file matches `.env*` (other than
        `.env.template`).
      - (`SECURITY.md` itself moved to its own item below.)

- [ ] **Stricter convention enforcement + `jwa-tobrew test`**
      Two strands:
      1. **Convention enforcement** — `jwa-tobrew doctor` should fail (not
         warn) if any `.rb` violates conventions: missing `desc`, missing
         `homepage`, version not semver, source URL not a github.com URL,
         filename != cask/formula `name`. Same checks run in CI. The
         conventions live with the security ground rules; cross-link.
      2. **`jwa-tobrew test`** — release a candidate without making it the
         "canonical" entry. Open question: should test:
            (a) cut a GitHub release marked `--prerelease` + write an `.rb`
                to a separate dir (e.g. `Test/`) that brew can install
                explicitly but doesn't surface in tap auto-completion;
            (b) commit to a `test/<name>` branch in the tap and never to
                main until `jwa-tobrew promote <name>`;
            (c) build the artifact, write a local `.rb`, and only suggest
                `brew install --formula <path>` (no GitHub release at all).
         (c) is simplest; (a) is closest to "real brew test"; (b) is most
         flexible. Decide before building.

- [ ] **Skills alignment follow-up**
      The active contract is ADR 0006 + ADR 0009: `agentskills` owns skill
      content and harness links, and `jwa-tobrew init` calls
      `agentskills bootstrap --skill release` instead of embedding a release
      skill template. Remaining work is to remove any obsolete generated skill
      output from local checkouts and keep tap docs pointed at the canonical
      skills in `jwa91/agentskills`.

- [ ] **Schema + strict typing, no defensive defaults**
      Inspired by prehandover's `schema.json` + `.golangci.yml`. The point is
      to make wrong state a compile-time or load-time error, not a silently
      degraded run. No backwards-compat shims — change the types, fix the
      callers, ship it.
      - Ship `schema.json` for `tap.toml` (JSON Schema 2020-12, mirroring
        prehandover's pattern). Add a CI job that validates the committed
        `tap.toml` against it, and reject unknown fields
        (`additionalProperties: false`).
      - Replace the `map[string]string` template payloads in `init.go` and
        `release.go::writeNewTapFile` with typed structs. `text/template`
        will then error on missing field references at parse time instead
        of rendering an empty string.
      - Audit `parseTapFile` and `LoadConfig` for "if missing, use empty"
        patterns. Required fields (e.g. `version`, `url`, `desc`,
        `homepage`) should fail loud, not produce an `Item{}` with empty
        strings that propagate into `tap.toml`. Pair with the doctor work
        in item #2 — same set of required fields, single source of truth.
      - Stop swallowing errors with `warn(...); continue` in
        `manifest.go::ScanTap`. A malformed `.rb` is a bug, not a warning;
        the run should fail and the offending file should be named.
      - Adopt prehandover's `.golangci.yml` (govet, staticcheck, errorlint,
        nilerr, ineffassign, copyloopvar, durationcheck, bodyclose,
        misspell). Gate CI on it. Fix every finding rather than nolinting.
      - If any of the above forces a CLI flag or struct field to change
        shape, just change it. No deprecation aliases. Update the in-repo
        skill and ADRs in the same commit.

- [ ] **`SECURITY.md` + repository hardening doc**
      Split out from item #1 because it's docs work, not CI/branch-policy
      work, and the doc is the spec the CI in #1 enforces against. Mirrors
      prehandover's `docs/repository-hardening.md` shape.
      - Write `SECURITY.md` at repo root: threat model (who can push, what
        a commit is allowed to contain, how compromise is detected),
        supported versions (only latest tag of each item), private
        disclosure channel, scope (this tap repo + each `cmd/` CLI inside
        it).
      - Write `docs/repository-hardening.md` enumerating the exact GitHub
        settings to enable (branch protection, required checks, signed
        commits, force-push block, default workflow token = read-only,
        first-time-contributor approval). This is the human-readable
        version of what item #1's CI/CODEOWNERS work enforces.
      - Add a `scripts/audit-hardening.sh` (or `jwa-tobrew doctor
        --remote`) that calls `gh api repos/jwa91/homebrew-tap` and
        reports drift between the documented expected settings and
        actual GitHub state. Read-only — no auto-apply.
      - Cross-link `SECURITY.md` ⇄ `docs/repository-hardening.md` ⇄
        `~/dotfiles/docs/security-ground-rules.md` ⇄ ADRs 0001/0002/0003.
      - Per "no backwards-compat": when this doc says HTTPS pushes are
        forbidden, the corresponding code path stays gone — do not add
        an opt-in flag to re-enable it.

---

## Done

- [x] Initial CLI: `add`, `bump`, `config`, `deps`, `doctor`, `init`,
      `release`, `upgrade`
- [x] Strip 1P / token-rotation / usage-log → moved to `jwa-harden` TODO
- [x] SSH-only for tap pushes
- [x] `tap.toml` + `tap.local.toml` manifests
- [x] In-repo skill (lives at `.claude/skills/jwa-tobrew/SKILL.md`)
- [x] `jwa-tobrew init` writes `.env.template`, `.gitignore` entries, and
      a project-level `release` skill into the target project
