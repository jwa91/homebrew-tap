---
name: tap-scaffolding
description: Modify or extend the scaffolding system inside `jwa-tobrew` (the templates that `init` writes into target projects). Trigger when the user says "add a new scaffold kind", "change what init writes", "update the templates", "add a new file to scaffolds", or asks how the embedded templates are wired. This skill is for *maintainers of jwa-tobrew*; end-users wanting to scaffold a project should use `scaffold-cli` (Go) or `jwa-tobrew init --kind=cask` (cask).
---

# How jwa-tobrew's scaffolding system works

`init` is the only command that writes files into target projects. Every scaffolded artifact has a corresponding contract enforced by `align` — that pairing is the system. If you change one without the other, `align` reports drift on freshly-scaffolded projects, which is the canary that something is misaligned.

## File layout

```
tools/jwa-tobrew/
├── init.go                    runs scaffold; auto-detects kind; prints per-kind hints
├── templates.go               //go:embed lines, one per template file
├── align.go                   the contract: alignProject() + alignTap()
└── templates/
    ├── goreleaser.yaml.tmpl   Go projects only
    ├── release.sh.tmpl        cask + formula projects (used to be two; unified in v0.3-ish)
    ├── cask.rb.tmpl           cask .rb shape (used by writeNewTapFile in release.go)
    ├── formula.rb.tmpl        formula .rb shape (single-asset; multi-platform formulas
                               come from source-repo GoReleaser, not from here)
    └── skill-release.md.tmpl  per-project release skill written by writeAuxiliary
```

Templates use Go's `text/template` syntax with double braces. To emit a literal `{{ ... }}` (e.g. for GoReleaser's own template variables in `goreleaser.yaml.tmpl`), wrap it: `{{`{{ .Version }}`}}`.

## The two write paths

Only two functions write template-rendered files to disk:

1. **`init.go::writeTemplate(path, tmpl, data, force)`** — renders a template and writes. Refuses to overwrite unless `force=true`. Used by `scaffold` and `writeAuxiliary` to drop scaffolded files into target projects.
2. **`release.go::writeNewTapFile`** — renders `cask.rb.tmpl` or `formula.rb.tmpl` and writes a brand-new tap entry. Always passes `force=true` because the file shouldn't have existed yet (`add` checks; `release` checks).

`scaffold` itself does kind dispatch:

```go
switch kind {
case "go":
    writeTemplate(".goreleaser.yaml", goreleaserTmpl, data, force)
case "cask", "formula":
    writeTemplate("scripts/release.sh", releaseShTmpl, data, force)
    chmod 0o755
}
writeAuxiliary(dir, data, force)  // .env.template, .gitignore, project skill
```

## Adding a new scaffold kind

Adding a new project kind (say `node`) is a 5-step recipe:

1. **Add a template file** at `tools/jwa-tobrew/templates/<name>.tmpl` (e.g. `release-node.sh.tmpl`).
2. **Add a `//go:embed` line** in `templates.go` and a `var nodeReleaseTmpl string`.
3. **Extend `init.go::detectKind`** to recognise the language marker (`package.json`, `Cargo.toml`, etc.) and return `"node"`.
4. **Extend `init.go::scaffold`** with a `case "node":` branch that calls `writeTemplate` for the new template, plus a hint in the trailing per-kind switch.
5. **Extend `align.go::alignProject`'s switch** with a `case "node":` that requires the same files (e.g. the new release script). This is what closes the loop: a freshly-scaffolded project must satisfy `align` immediately, otherwise the scaffold is incomplete.

Then update the `tap-alignment` skill if the convention is interesting, and add a CHANGELOG entry. Run `jwa-tobrew init --kind=node` in a scratch dir, followed by `jwa-tobrew align`, to verify the contract closes.

## Adding a new file to all scaffolds (regardless of kind)

If every project needs a new file (say `.editorconfig`):

1. Add the template at `templates/editorconfig.tmpl` and a `//go:embed` line.
2. Extend `init.go::writeAuxiliary` (not `scaffold`) with a `writeTemplate` call. `writeAuxiliary` is the place for shared-across-kinds artifacts.
3. Extend `align.go::alignProject` with the corresponding existence check, **outside** the per-kind switch.
4. Update `prek.toml`'s `jwa-tobrew-align` hook regex if the new file's path isn't already covered.
5. CHANGELOG + skill update if the file is non-obvious.

## The contract the system implements

> Anything `init` writes is something `align` enforces. Anything `align` enforces is something `init` writes (or the user is told what to run, when auto-create isn't safe).

This is what makes the system honest. If you add an `align` rule without a scaffold for it, users get findings they can't auto-fix. If you add a scaffold without an `align` rule, users can drift away from the convention silently. Always update both in the same commit, and verify by scaffolding into a scratch dir + running `align` immediately after.

## What is not scaffolded (intentionally)

- **`.goreleaser.yaml` for non-Go projects** — they don't use it.
- **`scripts/release.sh` for Go projects** — they don't use it; `goreleaser release --clean` is the entire pipeline locally and in CI.
- **GitHub Actions workflows** — out of scope for `init`. The `scaffold-cli` skill describes the recommended `release.yml` for tag-driven CI release.
- **`SECURITY.md`, `CHANGELOG.md`, `CODEOWNERS`** — repo-level files, not project-level. Owned by the repo's own conventions, not by `jwa-tobrew init`.
- **A `LICENSE`** — too opinionated to template; user picks.

## Verifying changes

1. `go build ./tools/jwa-tobrew && go vet ./...`
2. Scratch test: `mkdir /tmp/scratch-cli && cd /tmp/scratch-cli && git init && go mod init test && /path/to/jwa-tobrew init --kind=go --name=test --desc="test"`
3. `jwa-tobrew align` in the scratch dir — must report no drift
4. `prek run --all-files` in the tap repo — must pass
5. CHANGELOG + skill updates if user-visible
