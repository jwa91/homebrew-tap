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
      - Document the threat model in `SECURITY.md`: who can push, what's
        allowed in a commit, how a compromise is detected.

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

- [ ] **Skills alignment** — the in-repo skill at
      `.claude/skills/jwa-tobrew/SKILL.md` and the embedded release-skill
      template at `cmd/jwa-tobrew/templates/skill-release.md.tmpl` have
      drifted from the actual command set. Decide whether to keep the
      in-repo skill as one big "tap operations" skill or split it into
      focused per-command skills (`tap-add`, `tap-bump`, `tap-release`,
      `tap-deps`). Either way: rewrite to match the current commands,
      cross-reference the security ground rules, and verify both skills
      describe the same model (no contradictions).

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
