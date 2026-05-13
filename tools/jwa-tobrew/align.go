package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// alignFinding is one piece of drift between the current repo and the
// conventions jwa-tobrew expects. Apply may be nil for findings that
// require manual work (e.g. running `jwa-tobrew init` with flags).
type alignFinding struct {
	Path   string // repo-relative path the change touches
	Action string // one-line description of what would happen
	Apply  func() error
}

func runAlign(args []string) error {
	fs := subFlagSet("align", "report (or apply) drift from current jwa-tobrew conventions")
	apply := fs.Bool("apply", false, "actually make the changes (default: report only)")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	mode := detectAlignMode(cwd)
	var findings []alignFinding
	switch mode {
	case "tap":
		findings = alignTap(cwd)
	case "project":
		findings = alignProject(cwd)
	default:
		return errors.New("not a tap (no Casks/ + Formula/) and not a recognised project (no go.mod / Package.swift / *.xcodeproj)")
	}

	banner("align — %s mode", mode)
	if len(findings) == 0 {
		ok("no drift detected")
		return nil
	}
	for _, f := range findings {
		if f.Apply == nil {
			warn("%s — %s [manual]", f.Path, f.Action)
		} else {
			info("%s — %s", f.Path, f.Action)
		}
	}
	if !*apply {
		hint("run with --apply to make the auto-applyable changes")
		return nil
	}

	applied, manual, failed := 0, 0, 0
	for _, f := range findings {
		if f.Apply == nil {
			manual++
			continue
		}
		if err := f.Apply(); err != nil {
			fail("%s: %v", f.Path, err)
			failed++
			continue
		}
		ok("applied: %s", f.Path)
		applied++
	}
	fmt.Fprintf(os.Stderr, "\napplied=%d manual=%d failed=%d\n", applied, manual, failed)
	if failed > 0 {
		return fmt.Errorf("%d apply step(s) failed", failed)
	}
	return nil
}

func detectAlignMode(dir string) string {
	if exists(filepath.Join(dir, "Casks")) && exists(filepath.Join(dir, "Formula")) {
		return "tap"
	}
	if exists(filepath.Join(dir, "go.mod")) || exists(filepath.Join(dir, "Package.swift")) {
		return "project"
	}
	for _, pat := range []string{"*.xcodeproj", "*.xcworkspace"} {
		if m, _ := filepath.Glob(filepath.Join(dir, pat)); len(m) > 0 {
			return "project"
		}
	}
	return ""
}

// skillSymlinkFindings is the convention shared by tap and project repos:
// skills live in .agents/skills/<name>/, and .claude/skills/<name> is a
// symlink to ../../.agents/skills/<name>. This lets non-Claude agent
// runtimes (codex, cursor, etc.) symlink the same source from their own
// dirs without duplicating the skill content.
func skillSymlinkFindings(root, name string) []alignFinding {
	src := filepath.Join(root, ".agents", "skills", name)
	link := filepath.Join(root, ".claude", "skills", name)
	target := filepath.Join("..", "..", ".agents", "skills", name)

	srcExists := exists(src)
	linkInfo, linkErr := os.Lstat(link)
	linkExists := linkErr == nil
	linkIsSymlink := linkExists && linkInfo.Mode()&os.ModeSymlink != 0

	var out []alignFinding
	switch {
	case !srcExists && !linkExists:
		// Nothing to do — the skill simply doesn't exist here.
	case !srcExists && linkExists:
		// Migrate .claude/skills/<name> → .agents/skills/<name> + symlink.
		out = append(out, alignFinding{
			Path:   filepath.Join(".agents", "skills", name),
			Action: fmt.Sprintf("migrate skill from .claude/skills/%s to .agents/skills/%s and replace with symlink", name, name),
			Apply: func() error {
				if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
					return err
				}
				if err := os.Rename(link, src); err != nil {
					return err
				}
				return os.Symlink(target, link)
			},
		})
	case srcExists && !linkExists:
		out = append(out, alignFinding{
			Path:   filepath.Join(".claude", "skills", name),
			Action: fmt.Sprintf("create symlink → ../../.agents/skills/%s", name),
			Apply: func() error {
				if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
					return err
				}
				return os.Symlink(target, link)
			},
		})
	case srcExists && linkExists && !linkIsSymlink:
		out = append(out, alignFinding{
			Path:   filepath.Join(".claude", "skills", name),
			Action: fmt.Sprintf("replace dir with symlink to ../../.agents/skills/%s (existing dir will be removed)", name),
			Apply: func() error {
				if err := os.RemoveAll(link); err != nil {
					return err
				}
				return os.Symlink(target, link)
			},
		})
	}
	return out
}

func alignTap(dir string) []alignFinding {
	var f []alignFinding

	// ADR 0006: the tap-managed skills are materialised by
	// `make bootstrap-skills` (which invokes `agentskills bootstrap`),
	// not vendored. Drift here means the bootstrap mechanism is
	// missing, not that a specific directory is absent: the skill dirs
	// are generated-and-gitignored. Check that the CLI is reachable so
	// `make bootstrap-skills` can run; the user installs it via
	// `brew install jwa91/tap/agentskills`.
	if _, err := exec.LookPath("agentskills"); err != nil {
		f = append(f, alignFinding{
			Path:   ".agents/skills/",
			Action: "missing `agentskills` on PATH — install via `brew install jwa91/tap/agentskills`, then run `make bootstrap-skills`",
		})
	}

	if !exists(filepath.Join(dir, "docs", "adr")) {
		f = append(f, alignFinding{
			Path:   "docs/adr/",
			Action: "missing — create ADR directory and document intentional design choices",
		})
	}

	readme := filepath.Join(dir, "README.md")
	if exists(readme) {
		body, _ := os.ReadFile(readme)
		s := string(body)
		if !strings.Contains(s, readmeBeginMarker) || !strings.Contains(s, readmeEndMarker) {
			f = append(f, alignFinding{
				Path:   "README.md",
				Action: fmt.Sprintf("missing %q … %q markers — wrap the items section to enable auto-update", readmeBeginMarker, readmeEndMarker),
			})
		}
	}

	if !exists(filepath.Join(dir, "CHANGELOG.md")) {
		f = append(f, alignFinding{
			Path:   "CHANGELOG.md",
			Action: "missing — create a Keep-a-Changelog file to version this tap's evolution",
		})
	}

	// If the tap embeds a Go CLI (its own go.mod at root), it must
	// publish itself the same way every other Go CLI does — via
	// .goreleaser.yaml.
	if exists(filepath.Join(dir, "go.mod")) && !exists(filepath.Join(dir, ".goreleaser.yaml")) {
		f = append(f, alignFinding{
			Path:   ".goreleaser.yaml",
			Action: "missing — tap embeds a Go CLI; required to publish it back to the tap",
		})
	}

	// ADR 0006 — the boundary contract with agentskills must be on
	// disk. The tap's bootstrap mechanism (planned migration target)
	// and the `align` rules below derive from it, so a missing ADR
	// 0006 means the rules below have no anchor.
	if !exists(filepath.Join(dir, "docs", "adr", "0006-agentskills-boundary-contract.md")) {
		f = append(f, alignFinding{
			Path:   filepath.Join("docs", "adr", "0006-agentskills-boundary-contract.md"),
			Action: "missing — boundary contract with agentskills must be present (see ADR 0006 in this directory)",
		})
	}

	// ADR 0007 — the boundary contract with dotfiles. The
	// no-Brewfile and no-security-doc-duplication rules below derive
	// from it.
	if !exists(filepath.Join(dir, "docs", "adr", "0007-dotfiles-tap-contract.md")) {
		f = append(f, alignFinding{
			Path:   filepath.Join("docs", "adr", "0007-dotfiles-tap-contract.md"),
			Action: "missing — boundary contract with dotfiles must be present (see ADR 0007 in this directory)",
		})
	}

	// ADR 0007 forbidden pattern (c): the tap must not have its own
	// Brewfile. `~/dotfiles/Brewfile` is the single source of truth
	// for what gets installed on a fresh Mac.
	if exists(filepath.Join(dir, "Brewfile")) {
		f = append(f, alignFinding{
			Path:   "Brewfile",
			Action: "forbidden — Brewfile belongs in ~/dotfiles only (ADR 0007). Remove this file.",
		})
	}

	// ADR 0007 reference-loop rule: the security ground rules live
	// in ~/dotfiles/docs/security-ground-rules.md only. The tap must
	// not host a copy under any path.
	if hits := findSecurityGroundRulesDupes(dir); len(hits) > 0 {
		for _, hit := range hits {
			rel, err := filepath.Rel(dir, hit)
			if err != nil {
				rel = hit
			}
			f = append(f, alignFinding{
				Path:   rel,
				Action: "forbidden — security ground rules live in ~/dotfiles/docs/security-ground-rules.md (ADR 0007). Remove this duplicate and link to the canonical doc instead.",
			})
		}
	}
	return f
}

// findSecurityGroundRulesDupes walks the tap looking for any file
// named `security-ground-rules.md`. ADR 0007 forbids duplicating the
// dotfiles security model inside the tap. We skip VCS internals and
// any vendored agentskills bootstrap output so this stays cheap and
// stable across machines.
func findSecurityGroundRulesDupes(root string) []string {
	var hits []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if path != root && (base == ".git" || base == "node_modules" || base == "vendor" || base == ".cache") {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == "security-ground-rules.md" {
			hits = append(hits, path)
		}
		return nil
	})
	return hits
}

func alignProject(dir string) []alignFinding {
	var f []alignFinding

	envT := filepath.Join(dir, ".env.template")
	if !exists(envT) {
		f = append(f, alignFinding{
			Path:   ".env.template",
			Action: "create with op:// references for $GITHUB_TOKEN/$GH_TOKEN",
			Apply:  func() error { return writeEnvTemplate(envT, false) },
		})
	}

	gi := filepath.Join(dir, ".gitignore")
	body, _ := os.ReadFile(gi)
	if !gitignoreHas(string(body), ".env") {
		f = append(f, alignFinding{
			Path:   ".gitignore",
			Action: "add `.env` block (`.env`, `.env.local`, `.env.*.local`)",
			Apply:  func() error { return ensureGitignoreEntry(gi, ".env") },
		})
	}

	f = append(f, skillSymlinkFindings(dir, "release")...)

	// Per-kind release contract:
	//   Go             → .goreleaser.yaml (no shell wrapper; goreleaser is
	//                    the pipeline — same for local and CI).
	//   Cask / Formula → scripts/release.sh (wraps `jwa-tobrew release`).
	// detectKind only returns "go" or "cask"; formula has no file marker
	// so its contract isn't enforced here (would be in a future doctor
	// pass that reads the .rb file in the tap).
	switch detectKind(dir) {
	case "go":
		if !exists(filepath.Join(dir, ".goreleaser.yaml")) {
			f = append(f, alignFinding{
				Path:   ".goreleaser.yaml",
				Action: "missing — required for all Go CLIs (run `jwa-tobrew init --kind=go`)",
			})
		}
	case "cask":
		if !exists(filepath.Join(dir, "scripts", "release.sh")) {
			f = append(f, alignFinding{
				Path:   "scripts/release.sh",
				Action: "missing — run `jwa-tobrew init --kind=cask` to scaffold",
			})
		}
	}
	return f
}

func gitignoreHas(body, want string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}
