package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func runBump(args []string) error {
	fs := subFlagSet("bump", "re-sync an existing Cask or Formula to a published GitHub release")
	asset := fs.String("asset", "", "exact release asset filename (overrides auto-detection)")
	dryRun := fs.Bool("dry-run", false, "compute the change but do not write or commit")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) < 1 {
		return errors.New("usage: brewtap bump <name> [version]")
	}
	name := rest[0]
	wantVersion := ""
	if len(rest) >= 2 {
		wantVersion = strings.TrimPrefix(rest[1], "v")
	}

	c := LoadConfig()
	tap, err := TapDir(c)
	if err != nil {
		return err
	}
	art, err := FindArtifact(tap, name)
	if err != nil {
		return err
	}
	full := filepath.Join(tap, art.RelPath())
	body, err := os.ReadFile(full)
	if err != nil {
		return err
	}

	owner, repo, err := parseGitHubURL(string(body))
	if err != nil {
		return err
	}
	info("found %s at %s (source: %s/%s)", art.Kind, art.RelPath(), owner, repo)

	rel, err := LatestRelease(owner, repo)
	if err != nil {
		return err
	}
	tagVer := strings.TrimPrefix(rel.TagName, "v")
	if wantVersion == "" {
		wantVersion = tagVer
		info("latest release is %s", rel.TagName)
	} else if wantVersion != tagVer {
		warn("requested version %s but latest published release is %s", wantVersion, rel.TagName)
	}

	// Pick the asset URL: explicit override, single asset, or pattern match.
	dlURL := ""
	switch {
	case *asset != "":
		for _, a := range rel.Assets {
			if a.Name == *asset {
				dlURL = a.BrowserDownloadURL
				break
			}
		}
		if dlURL == "" {
			return fmt.Errorf("asset %q not found in release %s", *asset, rel.TagName)
		}
	default:
		want := substituteVersion(string(body), wantVersion)
		dlURL = matchAssetURL(rel, want)
		if dlURL == "" {
			if len(rel.Assets) == 1 {
				dlURL = rel.Assets[0].BrowserDownloadURL
			} else {
				return fmt.Errorf("could not determine asset for version %s — pass --asset <name>", wantVersion)
			}
		}
	}
	info("hashing %s", dlURL)
	sha, err := Sha256OfURL(dlURL)
	if err != nil {
		return err
	}

	if *dryRun {
		ok("would set version=%s sha256=%s in %s", wantVersion, sha, art.RelPath())
		return nil
	}

	changed, err := UpdateVersionAndSha(full, wantVersion, sha)
	if err != nil {
		return err
	}
	if !changed {
		ok("%s already at version %s with matching sha — nothing to do", art.RelPath(), wantVersion)
		return nil
	}
	ok("updated %s to %s", art.RelPath(), wantVersion)

	hash, err := GitCommitAndPush(c, tap, []string{art.RelPath()}, fmt.Sprintf("Update %s to %s", name, wantVersion))
	if err != nil {
		return err
	}
	if hash != "" {
		ok("pushed commit %s", shortHash(hash))
	}
	return nil
}

// parseGitHubURL extracts owner/repo from the first github.com URL in body.
func parseGitHubURL(body string) (string, string, error) {
	re := regexp.MustCompile(`github\.com/([^/\s"]+)/([^/\s"]+)`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		return "", "", errors.New("no github.com URL found in tap file")
	}
	owner, repo := m[1], strings.TrimSuffix(m[2], ".git")
	return owner, repo, nil
}

// substituteVersion replaces #{version} interpolations in the existing url
// line so we can predict the asset filename.
func substituteVersion(body, version string) string {
	re := regexp.MustCompile(`#\{version\}`)
	return re.ReplaceAllString(body, version)
}

// matchAssetURL tries to find the asset whose name appears in body.
func matchAssetURL(rel *GHRelease, body string) string {
	for _, a := range rel.Assets {
		if strings.Contains(body, "/"+a.Name) {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

func shortHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}
