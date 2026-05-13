package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// classNameFor turns "my-tool" into "MyTool" for Formula class naming.
func classNameFor(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		if p == "" {
			continue
		}
		runes := []rune(p)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}

func runRelease(args []string) error {
	fs := subFlagSet("release", "tag, create GitHub release, and update the tap (run inside a project repo)")
	kind := fs.String("kind", "", "release kind: go | cask | formula (auto-detected if omitted)")
	version := fs.String("version", "", "version to release (e.g. 0.5.0). Required for non-Go kinds")
	asset := fs.String("asset", "", "path to the artifact to upload (required for cask/formula)")
	name := fs.String("name", "", "tap artifact name (defaults to repo name)")
	desc := fs.String("desc", "", "tap entry description (cask/formula, only used on first publish)")
	push := fs.Bool("push", true, "push the tag to origin before creating the release")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	c := LoadConfig()
	if err := requireGitHubToken(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if !isGitRepo(cwd) {
		return errors.New("current directory is not a git repository")
	}
	if dirty, _ := gitDirty(cwd); dirty {
		return errors.New("working tree has uncommitted changes — commit or stash first")
	}
	owner, repo, err := repoOriginInfo(cwd)
	if err != nil {
		return err
	}
	if *name == "" {
		*name = repo
	}
	if *kind == "" {
		*kind = detectKind(cwd)
	}

	switch *kind {
	case "go":
		return releaseGo(c, cwd, *version, *push)
	case "cask", "formula":
		if *version == "" || *asset == "" {
			return fmt.Errorf("--version and --asset are required for kind=%s", *kind)
		}
		return releaseBinary(c, cwd, owner, repo, *kind, *name, *version, *asset, *desc, *push)
	default:
		return errors.New("could not auto-detect kind — pass --kind=go|cask|formula")
	}
}

// releaseGo defers the entire flow to GoReleaser, which expects a tag to exist.
// $GITHUB_TOKEN must already be in env (provided by jwa-harden run).
func releaseGo(_ Config, dir, version string, push bool) error {
	if _, err := exec.LookPath("goreleaser"); err != nil {
		return errors.New("goreleaser not installed — `brew install goreleaser`")
	}
	if version != "" {
		tag := "v" + strings.TrimPrefix(version, "v")
		if err := tagAndPush(dir, tag, push); err != nil {
			return err
		}
	} else if !hasHeadTag(dir) {
		return errors.New("HEAD is not tagged — pass --version=X.Y.Z or tag manually before releasing")
	}
	info("running goreleaser release --clean")
	cmd := exec.Command("goreleaser", "release", "--clean")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func releaseBinary(c Config, dir, owner, repo, kind, name, version, asset, desc string, push bool) error {
	if _, err := os.Stat(asset); err != nil {
		return fmt.Errorf("asset not found: %s", asset)
	}
	tag := "v" + strings.TrimPrefix(version, "v")
	ver := strings.TrimPrefix(version, "v")

	if err := tagAndPush(dir, tag, push); err != nil {
		return err
	}

	info("creating GitHub release %s with auto-generated notes", tag)
	if err := CreateRelease(tag, tag, "", []string{asset}); err != nil {
		return err
	}
	dlURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		owner, repo, tag, filepath.Base(asset))
	info("hashing uploaded asset")
	sha, err := Sha256OfFile(asset)
	if err != nil {
		return err
	}

	tap, err := TapDir(c)
	if err != nil {
		return err
	}
	art := Artifact{Kind: kind, Name: name}
	full := filepath.Join(tap, art.RelPath())
	if !exists(full) {
		info("creating new %s at %s", kind, art.RelPath())
		if err := writeNewTapFile(full, kind, name, owner, repo, ver, sha, desc, dlURL); err != nil {
			return err
		}
	} else {
		body, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		if n := AssetCount(body); n > 1 {
			return fmt.Errorf("%s has %d url lines (multi-platform); jwa-tobrew can't safely update it — the source repo's release pipeline owns this formula",
				art.RelPath(), n)
		}
		changed, err := UpdateVersionAndSha(full, ver, sha)
		if err != nil {
			return err
		}
		if !changed {
			ok("%s already at %s — nothing to commit", art.RelPath(), ver)
			return nil
		}
	}
	if err := writeManifests(c, tap); err != nil {
		warn("could not refresh manifests/README: %v", err)
	}
	hash, err := GitCommitAndPush(c, tap, []string{art.RelPath(), "tap.toml", "README.md"}, fmt.Sprintf("Update %s to %s", name, ver))
	if err != nil {
		return err
	}
	if hash != "" {
		ok("pushed tap commit %s — install with `brew install %s/%s/%s`", shortHash(hash), c.TapOwner, c.TapName, name)
	}
	return nil
}

func writeNewTapFile(path, kind, name, owner, repo, version, sha, desc, dlURL string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Convert the concrete download URL into a templated form by replacing
	// every occurrence of the version that sits between two of the common
	// separators (/, _, -, .) used in GitHub release paths and asset filenames.
	verRe := regexp.MustCompile(`([/_-])` + regexp.QuoteMeta(version) + `([/_.])`)
	urlTmpl := verRe.ReplaceAllString(dlURL, `${1}#{version}${2}`)
	data := map[string]string{
		"Name":      name,
		"ClassName": classNameFor(name),
		"Owner":     owner,
		"Repo":      repo,
		"Version":   version,
		"Sha256":    sha,
		"Desc":      coalesce(desc, "TODO: one-line description"),
		"URL":       urlTmpl,
		"Asset":     filepath.Base(dlURL),
	}
	tmpl := caskRbTmpl
	if kind == "formula" {
		tmpl = formulaRbTmpl
	}
	return writeTemplate(path, tmpl, data, true)
}
