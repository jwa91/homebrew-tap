package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// makeTapSkeleton creates a minimal tap-shaped directory under t.TempDir
// that detectAlignMode classifies as "tap": it has Casks/ and Formula/
// directories. Tap-managed skills are no longer vendored (ADR 0006 +
// migration plan) — they're materialised by `make bootstrap-skills`,
// so no skill dirs or symlinks are created here. The new
// agentskills-on-PATH finding is exercised separately.
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

// TestAlignTap_NoJwaTobrewSkillSymlinkFinding verifies that after the
// agentskills migration, alignTap no longer expects a vendored
// .agents/skills/jwa-tobrew/ tree or its .claude/skills/jwa-tobrew
// symlink. Those are now generated artifacts of `make bootstrap-skills`.
func TestAlignTap_NoJwaTobrewSkillSymlinkFinding(t *testing.T) {
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	for _, f := range findings {
		if strings.Contains(f.Path, filepath.Join(".agents", "skills", "jwa-tobrew")) ||
			strings.Contains(f.Path, filepath.Join(".claude", "skills", "jwa-tobrew")) {
			t.Fatalf("unexpected legacy skill finding after migration: %+v", f)
		}
	}
}

// TestAlignTap_AgentskillsOnPATH_Satisfied is a smoke test: when the
// `agentskills` binary is available on PATH (true in this dev
// environment per the migration pre-gates), alignTap must not emit the
// install-agentskills finding.
func TestAlignTap_AgentskillsOnPATH_Satisfied(t *testing.T) {
	// Only meaningful if PATH actually contains agentskills; skip on
	// hosts where it isn't installed (CI without the binary).
	if _, err := exec.LookPath("agentskills"); err != nil {
		t.Skip("agentskills not on PATH; skipping satisfied-case smoke test")
	}
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	for _, f := range findings {
		if strings.Contains(f.Action, "agentskills") {
			t.Fatalf("did not expect agentskills-on-PATH finding when binary is installed: %+v", f)
		}
	}
}

// TestAlignTap_AgentskillsOnPATH_Missing forces the lookup to fail by
// setting PATH to a directory that cannot contain `agentskills`, then
// asserts the finding fires with the expected manual install hint.
func TestAlignTap_AgentskillsOnPATH_Missing(t *testing.T) {
	emptyDir := t.TempDir()
	t.Setenv("PATH", emptyDir)
	dir := makeTapSkeleton(t)
	findings := alignTap(dir)
	var hit *alignFinding
	for i := range findings {
		if strings.Contains(findings[i].Action, "agentskills") {
			hit = &findings[i]
			break
		}
	}
	if hit == nil {
		t.Fatalf("expected agentskills-on-PATH finding, got: %v", findings)
	}
	if !strings.Contains(hit.Action, "brew install jwa91/tap/agentskills") {
		t.Errorf("finding should suggest brew install: %q", hit.Action)
	}
	if hit.Apply != nil {
		t.Errorf("agentskills-on-PATH finding must be manual-only (Apply == nil)")
	}
}
