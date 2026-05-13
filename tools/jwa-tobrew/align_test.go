package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeTapSkeleton creates a minimal tap-shaped directory under t.TempDir
// that detectAlignMode classifies as "tap": it has Casks/ and Formula/
// directories, plus the .agents/skills/jwa-tobrew + .claude/skills
// symlink the existing skillSymlinkFindings logic already requires.
// Tests then add or remove specific files to assert the new findings.
func makeTapSkeleton(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	must(os.MkdirAll(filepath.Join(dir, "Casks"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, "Formula"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, "docs", "adr"), 0o755))
	// jwa-tobrew skill source + symlink so skillSymlinkFindings stays
	// quiet; the four new findings are what we want to isolate.
	must(os.MkdirAll(filepath.Join(dir, ".agents", "skills", "jwa-tobrew"), 0o755))
	must(os.MkdirAll(filepath.Join(dir, ".claude", "skills"), 0o755))
	must(os.Symlink(
		filepath.Join("..", "..", ".agents", "skills", "jwa-tobrew"),
		filepath.Join(dir, ".claude", "skills", "jwa-tobrew"),
	))
	// CHANGELOG + README with markers so those findings stay quiet.
	must(os.WriteFile(filepath.Join(dir, "CHANGELOG.md"), []byte("# Changelog\n"), 0o644))
	must(os.WriteFile(
		filepath.Join(dir, "README.md"),
		[]byte("# tap\n"+readmeBeginMarker+"\n"+readmeEndMarker+"\n"),
		0o644,
	))
	// Both ADRs present so they aren't flagged unless a test removes
	// them on purpose.
	must(os.WriteFile(filepath.Join(dir, "docs", "adr", "0006-agentskills-boundary-contract.md"), []byte("# 0006\n"), 0o644))
	must(os.WriteFile(filepath.Join(dir, "docs", "adr", "0007-dotfiles-tap-contract.md"), []byte("# 0007\n"), 0o644))
	return dir
}

func findingFor(findings []alignFinding, pathSubstr string) *alignFinding {
	for i := range findings {
		if strings.Contains(findings[i].Path, pathSubstr) {
			return &findings[i]
		}
	}
	return nil
}

func TestAlignTap_ADR0006MustExist(t *testing.T) {
	dir := makeTapSkeleton(t)
	if err := os.Remove(filepath.Join(dir, "docs", "adr", "0006-agentskills-boundary-contract.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	findings := alignTap(dir)
	f := findingFor(findings, "0006-agentskills-boundary-contract.md")
	if f == nil {
		t.Fatalf("expected ADR 0006 finding, got: %v", findings)
	}
	if !strings.Contains(f.Action, "boundary contract with agentskills") {
		t.Errorf("unexpected action: %q", f.Action)
	}
	if f.Apply != nil {
		t.Errorf("ADR 0006 finding must be manual-only (Apply == nil)")
	}
}

func TestAlignTap_ADR0006Satisfied(t *testing.T) {
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	if f := findingFor(findings, "0006-agentskills-boundary-contract.md"); f != nil {
		t.Fatalf("did not expect ADR 0006 finding, got %+v", *f)
	}
}

func TestAlignTap_ADR0007MustExist(t *testing.T) {
	dir := makeTapSkeleton(t)
	if err := os.Remove(filepath.Join(dir, "docs", "adr", "0007-dotfiles-tap-contract.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	findings := alignTap(dir)
	f := findingFor(findings, "0007-dotfiles-tap-contract.md")
	if f == nil {
		t.Fatalf("expected ADR 0007 finding, got: %v", findings)
	}
	if !strings.Contains(f.Action, "boundary contract with dotfiles") {
		t.Errorf("unexpected action: %q", f.Action)
	}
	if f.Apply != nil {
		t.Errorf("ADR 0007 finding must be manual-only (Apply == nil)")
	}
}

func TestAlignTap_ADR0007Satisfied(t *testing.T) {
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	if f := findingFor(findings, "0007-dotfiles-tap-contract.md"); f != nil {
		t.Fatalf("did not expect ADR 0007 finding, got %+v", *f)
	}
}

func TestAlignTap_BrewfileForbidden(t *testing.T) {
	dir := makeTapSkeleton(t)
	if err := os.WriteFile(filepath.Join(dir, "Brewfile"), []byte("tap \"jwa91/tap\"\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	findings := alignTap(dir)
	f := findingFor(findings, "Brewfile")
	if f == nil {
		t.Fatalf("expected Brewfile finding, got: %v", findings)
	}
	if !strings.Contains(f.Action, "~/dotfiles") {
		t.Errorf("Brewfile finding should point at ~/dotfiles: %q", f.Action)
	}
	if f.Apply != nil {
		t.Errorf("Brewfile finding must be manual-only (no auto-delete of user files)")
	}
}

func TestAlignTap_BrewfileSatisfied(t *testing.T) {
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	if f := findingFor(findings, "Brewfile"); f != nil {
		t.Fatalf("did not expect Brewfile finding, got %+v", *f)
	}
}

func TestAlignTap_SecurityGroundRulesForbidden(t *testing.T) {
	dir := makeTapSkeleton(t)
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	dupe := filepath.Join(dir, "docs", "security-ground-rules.md")
	if err := os.WriteFile(dupe, []byte("# dupe\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	findings := alignTap(dir)
	f := findingFor(findings, "security-ground-rules.md")
	if f == nil {
		t.Fatalf("expected security-ground-rules.md finding, got: %v", findings)
	}
	if !strings.Contains(f.Action, "~/dotfiles/docs/security-ground-rules.md") {
		t.Errorf("finding should point at canonical dotfiles location: %q", f.Action)
	}
	if f.Apply != nil {
		t.Errorf("security-ground-rules finding must be manual-only")
	}
}

func TestAlignTap_SecurityGroundRulesSatisfied(t *testing.T) {
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	if f := findingFor(findings, "security-ground-rules.md"); f != nil {
		t.Fatalf("did not expect security-ground-rules finding, got %+v", *f)
	}
}
