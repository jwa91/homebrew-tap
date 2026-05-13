package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// parseRepoRef accepts any reasonable spelling of a GitHub repo reference and
// returns owner, repo. Tolerates trailing path segments (e.g. /tree/main),
// trailing slashes, and a `.git` suffix.
//
// Recognised forms:
//
//	owner/repo
//	github.com/owner/repo
//	http(s)://github.com/owner/repo[.git][/...]
//	git@github.com:owner/repo[.git]
//	ssh://git@github.com/owner/repo[.git]
func parseRepoRef(s string) (owner, repo string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", errors.New("empty repo reference")
	}
	// Peel off scheme/host prefixes in priority order.
	for _, p := range []string{
		"git@github.com:",
		"ssh://git@github.com/",
		"https://github.com/",
		"http://github.com/",
		"github.com/",
	} {
		if rest, ok := strings.CutPrefix(s, p); ok {
			s = rest
			break
		}
	}
	// Reject other hosts loudly so a typo doesn't silently succeed.
	if strings.Contains(s, "://") || strings.Contains(s, "@") {
		return "", "", fmt.Errorf("only github.com refs are supported (got %q)", s)
	}
	s = strings.TrimSuffix(s, "/")
	parts := strings.Split(s, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected owner/repo, got %q", s)
	}
	return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

// GHRelease is the subset of GitHub release fields we use.
type GHRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// LatestRelease fetches the latest release for owner/repo via gh CLI.
// gh handles its own auth (env GH_TOKEN/GITHUB_TOKEN, or its keychain).
func LatestRelease(owner, repo string) (*GHRelease, error) {
	return ghRelease(fmt.Sprintf("repos/%s/%s/releases/latest", owner, repo))
}

// ReleaseByTag fetches a specific release by its tag (e.g. "v0.4.1").
func ReleaseByTag(owner, repo, tag string) (*GHRelease, error) {
	return ghRelease(fmt.Sprintf("repos/%s/%s/releases/tags/%s", owner, repo, tag))
}

func ghRelease(apiPath string) (*GHRelease, error) {
	out, err := exec.Command("gh", "api", apiPath).Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("gh api %s: %s", apiPath, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	var r GHRelease
	if err := json.Unmarshal(out, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateRelease creates a GitHub release and uploads the given asset paths.
// If notesFile is empty, GitHub auto-generates the changelog.
// $GITHUB_TOKEN must be in env (caller invokes requireGitHubToken first).
func CreateRelease(tag, title, notesFile string, assets []string) error {
	args := []string{"release", "create", tag, "--title", title}
	if notesFile != "" {
		args = append(args, "--notes-file", notesFile)
	} else {
		args = append(args, "--generate-notes")
	}
	args = append(args, assets...)
	cmd := exec.Command("gh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Sha256OfURL streams a URL and returns its hex sha256.
func Sha256OfURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	h := sha256.New()
	if _, err := io.Copy(h, resp.Body); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Sha256OfFile returns the hex sha256 of a local file.
func Sha256OfFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
