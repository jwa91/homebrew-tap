.PHONY: build install dev bootstrap bootstrap-skills setup clean hooks

BIN_DIR ?= $(HOME)/.local/bin

build:
	go build -o bin/jwa-tobrew ./tools/jwa-tobrew

install: build
	install -d $(BIN_DIR)
	install -m 0755 bin/jwa-tobrew $(BIN_DIR)/jwa-tobrew
	@echo "installed jwa-tobrew to $(BIN_DIR)/jwa-tobrew"
	@echo "ensure $(BIN_DIR) is on your PATH"

dev: install

# First-time release: tag, push, run goreleaser. After this, future releases
# of jwa-tobrew can use `jwa-harden run -- ./bin/jwa-tobrew release --kind go`.
# This bootstrap target uses `op run` directly because jwa-harden doesn't
# exist yet — replace once it does.
bootstrap: build
	@which goreleaser >/dev/null 2>&1 || (echo "install goreleaser: brew install goreleaser" && exit 1)
	@test -n "$(VERSION)" || (echo "usage: make bootstrap VERSION=0.1.0" && exit 2)
	@op whoami >/dev/null || (echo "1Password not signed in: op signin" && exit 1)
	git tag -a v$(VERSION) -m "v$(VERSION)" || true
	git push origin v$(VERSION)
	op run --env-file=.env.template -- goreleaser release --clean
	@echo "now pull the auto-committed Formula: git pull --rebase"

# Pinned ref the tap installs from. Policy (ADR 0006 + migration plan):
# commit a SHA from jwa91/agentskills's main branch. Bumping the pin is
# an explicit, reviewable commit in this tap. `main` is permitted via
# env override (`AGENTSKILLS_REF=main make bootstrap-skills`) for local
# development, but the committed default must never be `main` — that
# would re-create the dual-source-of-truth drift ADR 0006 exists to
# prevent. Per-skill tags (`<skill>/vX.Y.Z`) cannot compose into a
# single `--ref` call (agentskills bootstrap takes one ref).
AGENTSKILLS_REF ?= 155bd475f9c608c2f9240f0dfb20f3c102211f1a

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
# tap-managed skills. Does NOT depend on `build` — Go builds must not
# require network access.
setup: hooks bootstrap-skills

clean:
	rm -rf bin dist

# Install prek git hooks. One-time per clone. Hooks run on every commit
# thereafter; bypass with `git commit --no-verify` (do not, except for
# WIP branches you'll force-clean before pushing).
hooks:
	@which prek >/dev/null 2>&1 || (echo "install prek: brew install prek" && exit 1)
	prek install
	@echo "prek hooks installed. Run \`prek run --all-files\` to dry-run them."
