package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// TapDir resolves the homebrew-tap repo path. Honours $BREWTAP_DIR (via Config),
// otherwise probes a few common locations.
func TapDir(c Config) (string, error) {
	if c.TapDir != "" {
		if _, err := os.Stat(filepath.Join(c.TapDir, "Casks")); err == nil {
			return c.TapDir, nil
		}
		return "", fmt.Errorf("BREWTAP_DIR=%s does not look like a tap (missing Casks/)", c.TapDir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(home, "developer", "homebrew-tap"),
		filepath.Join(home, "Developer", "homebrew-tap"),
		filepath.Join(home, "src", "homebrew-tap"),
		filepath.Join(home, "code", "homebrew-tap"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "Casks")); err == nil {
			return c, nil
		}
	}
	return "", errors.New("could not locate homebrew-tap clone — set $BREWTAP_DIR to its path")
}

// Artifact selects the right tap subdirectory and file extension.
type Artifact struct {
	Kind string // "cask" or "formula"
	Name string
}

func (a Artifact) RelPath() string {
	if a.Kind == "cask" {
		return filepath.Join("Casks", a.Name+".rb")
	}
	return filepath.Join("Formula", a.Name+".rb")
}

// FindArtifact locates an existing cask or formula in the tap by name.
func FindArtifact(tap, name string) (*Artifact, error) {
	for _, k := range []string{"cask", "formula"} {
		a := Artifact{Kind: k, Name: name}
		if _, err := os.Stat(filepath.Join(tap, a.RelPath())); err == nil {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("no Cask or Formula named %q in %s", name, tap)
}

var (
	versionRE = regexp.MustCompile(`(?m)^(\s*version\s+)"[^"]+"`)
	sha256RE  = regexp.MustCompile(`(?m)^(\s*sha256\s+)"[^"]+"`)
)

// UpdateVersionAndSha rewrites the first version + sha256 lines in a tap file.
// Returns true if the file was changed.
func UpdateVersionAndSha(path, version, sha string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	orig := string(b)
	updated := versionRE.ReplaceAllString(orig, `${1}"`+version+`"`)
	updated = sha256RE.ReplaceAllString(updated, `${1}"`+sha+`"`)
	if updated == orig {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(updated), 0o644)
}

// GitCommitAndPush stages files in the tap and pushes to origin.
// Returns the commit hash, or "" if there was nothing to commit.
//
// Push goes through plain git, relying on the tap's `origin` remote being
// SSH (per the security ground rules). HTTPS origins are rejected with a
// hint to switch — wrapping git push with $GH_TOKEN would work but it
// violates "least surprise" and crosses the actor boundary (the human
// triggers the push; tokens are for automation).
func GitCommitAndPush(c Config, tap string, relPaths []string, message string) (string, error) {
	if err := requireSSHOrigin(tap); err != nil {
		return "", err
	}
	run := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tap
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
		}
		return strings.TrimSpace(string(out)), nil
	}
	for _, p := range relPaths {
		if _, err := run("add", p); err != nil {
			return "", err
		}
	}
	statusArgs := append([]string{"status", "--porcelain"}, relPaths...)
	if status, _ := run(statusArgs...); status == "" {
		return "", nil
	}
	if _, err := run("commit", "-m", message); err != nil {
		return "", err
	}
	hash, _ := run("rev-parse", "HEAD")

	push := exec.Command("git", "push")
	push.Dir = tap
	push.Stdout = os.Stdout
	push.Stderr = os.Stderr
	if err := push.Run(); err != nil {
		return hash, fmt.Errorf("git push: %w", err)
	}
	return hash, nil
}

// requireSSHOrigin returns nil iff the tap's origin remote is SSH.
func requireSSHOrigin(tap string) error {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = tap
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("tap has no `origin` remote — set one with `git -C %s remote add origin git@github.com:OWNER/REPO.git`", tap)
	}
	url := strings.TrimSpace(string(out))
	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		return nil
	}
	return fmt.Errorf("tap origin is %q but brwtp pushes over SSH only — switch with `git -C %s remote set-url origin git@github.com:OWNER/REPO.git`", url, tap)
}
