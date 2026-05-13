package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runAdd(args []string) error {
	fs := subFlagSet("add", "snapshot a published GitHub release into the tap (no token needed)")
	kind := fs.String("kind", "", "tap entry kind: cask | formula (auto-detected from asset extension)")
	name := fs.String("name", "", "tap entry name (defaults to the source repo name)")
	version := fs.String("version", "", "release version to snapshot (defaults to latest)")
	asset := fs.String("asset", "", "exact asset filename to use (defaults to single asset, or DMG-for-cask heuristic)")
	desc := fs.String("desc", "", "one-line description for the tap entry")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	rest := fs.Args()
	if len(rest) < 1 {
		return errors.New("usage: jwa-tobrew add <github-url> [flags]")
	}
	owner, repo, err := parseRepoRef(rest[0])
	if err != nil {
		return err
	}
	if *name == "" {
		*name = repo
	}

	c := LoadConfig()
	tap, err := TapDir(c)
	if err != nil {
		return err
	}

	// Fetch the release.
	var rel *GHRelease
	if *version == "" {
		rel, err = LatestRelease(owner, repo)
		if err != nil {
			return err
		}
		*version = strings.TrimPrefix(rel.TagName, "v")
		info("latest release: %s", rel.TagName)
	} else {
		tag := "v" + strings.TrimPrefix(*version, "v")
		rel, err = ReleaseByTag(owner, repo, tag)
		if err != nil {
			return err
		}
	}
	if len(rel.Assets) == 0 {
		return fmt.Errorf("release %s has no assets to snapshot", rel.TagName)
	}

	// Pick the asset.
	chosen := pickAsset(rel, *asset, *kind)
	if chosen == "" {
		listAssets(rel)
		return errors.New("could not auto-pick an asset — pass --asset <name>")
	}
	dlURL := assetURL(rel, chosen)
	if *kind == "" {
		*kind = inferKindFromAsset(chosen)
	}
	if *kind != "cask" && *kind != "formula" {
		return fmt.Errorf("could not infer kind from %q — pass --kind=cask|formula", chosen)
	}

	info("hashing %s", dlURL)
	sha, err := Sha256OfURL(dlURL)
	if err != nil {
		return err
	}

	art := Artifact{Kind: *kind, Name: *name}
	full := filepath.Join(tap, art.RelPath())
	if exists(full) {
		return fmt.Errorf("%s already exists in tap — use `brwtp bump %s` to update it", art.RelPath(), *name)
	}

	if err := writeNewTapFile(full, *kind, *name, owner, repo, *version, sha, *desc, dlURL); err != nil {
		return err
	}
	ok("wrote %s", art.RelPath())

	if err := writeManifests(c, tap); err != nil {
		warn("could not refresh manifests/README: %v", err)
	}

	hash, err := GitCommitAndPush(c, tap, []string{art.RelPath(), "tap.toml", "README.md"}, fmt.Sprintf("Add %s %s", *name, *version))
	if err != nil {
		return err
	}
	if hash != "" {
		ok("pushed commit %s — install with `brew install %s/%s/%s`", shortHash(hash), c.TapOwner, c.TapName, *name)
	}
	return nil
}

// assetExtsByKind lists the asset extensions we recognise for each tap kind,
// in priority order. Used both to pick an asset for a known kind and to infer
// the kind from an asset name (inverse lookup).
var assetExtsByKind = map[string][]string{
	"cask":    {".dmg", ".pkg", ".zip"},
	"formula": {".tar.gz", ".tgz"},
}

func pickAsset(rel *GHRelease, override, kind string) string {
	if override != "" {
		for _, a := range rel.Assets {
			if a.Name == override {
				return a.Name
			}
		}
		return ""
	}
	for _, ext := range assetExtsByKind[kind] {
		for _, a := range rel.Assets {
			if strings.HasSuffix(strings.ToLower(a.Name), ext) {
				return a.Name
			}
		}
	}
	if len(rel.Assets) == 1 {
		return rel.Assets[0].Name
	}
	return ""
}

func inferKindFromAsset(name string) string {
	n := strings.ToLower(name)
	for kind, exts := range assetExtsByKind {
		for _, ext := range exts {
			if strings.HasSuffix(n, ext) {
				return kind
			}
		}
	}
	return ""
}

func assetURL(rel *GHRelease, name string) string {
	for _, a := range rel.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

func listAssets(rel *GHRelease) {
	fmt.Fprintln(os.Stderr, "available assets:")
	for _, a := range rel.Assets {
		fmt.Fprintf(os.Stderr, "  %s\n", a.Name)
	}
}
