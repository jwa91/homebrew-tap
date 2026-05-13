.PHONY: bootstrap-skills setup hooks

# Pinned ref the tap installs from. Policy (ADR 0006 + migration plan):
# commit a SHA from jwa91/agentskills's main branch. Bumping the pin is
# an explicit, reviewable commit in this tap. `main` is permitted via
# env override (`AGENTSKILLS_REF=main make bootstrap-skills`) for local
# development, but the committed default must never be `main` — that
# would re-create the dual-source-of-truth drift ADR 0006 exists to
# prevent. Per-skill tags (`<skill>/vX.Y.Z`) cannot compose into a
# single `--ref` call (agentskills bootstrap takes one ref).
AGENTSKILLS_REF ?= b4976abc7dba164299fb6e869d15615a3d987cae

# Populate .agents/skills/ from the agentskills repo. Idempotent — safe
# to re-run on a fresh clone or after pulling. .agents/skills/ and
# .claude/skills/ are gitignored; this target is how they get created.
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

# Aggregate target for a fresh clone: install git hooks and materialise
# tap-managed skills. The CLIs themselves (jwa-tobrew, jwa-harden) are
# installed by `brew bundle install --file=~/dotfiles/Brewfile`, not by
# this Makefile — the tap is now a pure distribution channel (ADR 0008).
setup: hooks bootstrap-skills

# Install prek git hooks. One-time per clone. Hooks run on every commit
# thereafter; bypass with `git commit --no-verify` (do not, except for
# WIP branches you'll force-clean before pushing).
hooks:
	@which prek >/dev/null 2>&1 || (echo "install prek: brew install prek" && exit 1)
	prek install
	@echo "prek hooks installed. Run \`prek run --all-files\` to dry-run them."
