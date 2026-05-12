.PHONY: build install dev bootstrap clean

BIN_DIR ?= $(HOME)/.local/bin

build:
	go build -o bin/jwa-tobrew ./cmd/jwa-tobrew

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

clean:
	rm -rf bin dist
