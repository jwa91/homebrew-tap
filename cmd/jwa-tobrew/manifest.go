package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Item is one row in the tap manifest, derived from a Cask or Formula file.
type Item struct {
	Name     string   // matches the .rb basename
	Kind     string   // "cask" or "formula"
	Source   string   // e.g. github.com/jwa91/trnscrb
	Version  string   // current version pinned in the .rb
	URL      string   // the templated url line, with #{version}
	Asset    string   // basename of the URL asset
	Desc     string   // desc "..." line
	Homepage string   // homepage "..." line
	Deps     []string // depends_on / depends declarations as printable strings
}

// LocalAnnotation holds machine-specific info about an item.
type LocalAnnotation struct {
	Name       string
	LocalClone string // empty when no local clone was found
}

// ScanTap reads every .rb under Casks/ and Formula/ and returns one Item per file.
func ScanTap(tap string) ([]Item, error) {
	var items []Item
	for _, kind := range []string{"cask", "formula"} {
		dir := filepath.Join(tap, kindDir(kind))
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".rb") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			it, err := parseTapFile(path, kind)
			if err != nil {
				warn("skipping %s: %v", path, err)
				continue
			}
			items = append(items, it)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, nil
}

func kindDir(kind string) string {
	if kind == "cask" {
		return "Casks"
	}
	return "Formula"
}

var (
	versionLineRE  = regexp.MustCompile(`(?m)^\s*version\s+"([^"]+)"`)
	urlLineRE      = regexp.MustCompile(`(?m)^\s*url\s+"([^"]+)"`)
	descLineRE     = regexp.MustCompile(`(?m)^\s*desc\s+"([^"]+)"`)
	homepageLineRE = regexp.MustCompile(`(?m)^\s*homepage\s+"([^"]+)"`)

	// Captures `depends_on macos: ">= :tahoe"`, `depends_on "go"`,
	// `depends_on arch: :arm64`, and Formula-style `depends_on "git" => :build`.
	dependsOnRE = regexp.MustCompile(`(?m)^\s*depends_on\s+(.+)$`)
	// Captures Formula `depends "go"` (rare; older syntax).
	dependsRE = regexp.MustCompile(`(?m)^\s*depends\s+"([^"]+)"`)
)

func parseTapFile(path, kind string) (Item, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Item{}, err
	}
	name := strings.TrimSuffix(filepath.Base(path), ".rb")
	it := Item{Name: name, Kind: kind}
	if m := versionLineRE.FindStringSubmatch(string(body)); m != nil {
		it.Version = m[1]
	}
	if m := urlLineRE.FindStringSubmatch(string(body)); m != nil {
		it.URL = m[1]
		it.Asset = filepath.Base(m[1])
	}
	if m := descLineRE.FindStringSubmatch(string(body)); m != nil {
		it.Desc = m[1]
	}
	if m := homepageLineRE.FindStringSubmatch(string(body)); m != nil {
		it.Homepage = m[1]
	}
	if owner, repo, err := parseGitHubURL(string(body)); err == nil {
		it.Source = "github.com/" + owner + "/" + repo
	}
	for _, m := range dependsOnRE.FindAllStringSubmatch(string(body), -1) {
		if dep := tidyDep(m[1]); dep != "" {
			it.Deps = append(it.Deps, dep)
		}
	}
	for _, m := range dependsRE.FindAllStringSubmatch(string(body), -1) {
		it.Deps = append(it.Deps, m[1])
	}
	return it, nil
}

// tidyDep collapses whitespace in a depends_on argument list and trims
// trailing comments / empty fragments.
func tidyDep(s string) string {
	if i := strings.Index(s, "#"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// AnnotateLocal looks for a local clone of each item's source repo in the
// usual dev directories. It does not touch the network.
func AnnotateLocal(items []Item) []LocalAnnotation {
	out := make([]LocalAnnotation, 0, len(items))
	roots := devRoots()
	for _, it := range items {
		ann := LocalAnnotation{Name: it.Name}
		if it.Source != "" {
			parts := strings.Split(it.Source, "/")
			if len(parts) >= 3 {
				repoName := parts[len(parts)-1]
				for _, root := range roots {
					candidate := filepath.Join(root, repoName)
					if _, err := os.Stat(filepath.Join(candidate, ".git")); err == nil {
						ann.LocalClone = candidate
						break
					}
				}
			}
		}
		out = append(out, ann)
	}
	return out
}

func devRoots() []string {
	home, _ := os.UserHomeDir()
	if d := os.Getenv("DEV_DIR"); d != "" {
		return []string{d, filepath.Join(home, "developer"), filepath.Join(home, "Developer")}
	}
	return []string{
		filepath.Join(home, "Developer"),
		filepath.Join(home, "developer"),
		filepath.Join(home, "src"),
		filepath.Join(home, "code"),
	}
}

// writeManifests regenerates tap.toml + tap.local.toml under tap/.
func writeManifests(c Config, tap string) error {
	items, err := ScanTap(tap)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tap, "tap.toml"), []byte(renderTapTOML(c, items)), 0o644); err != nil {
		return err
	}
	annots := AnnotateLocal(items)
	if err := os.WriteFile(filepath.Join(tap, "tap.local.toml"), []byte(renderTapLocalTOML(items, annots)), 0o644); err != nil {
		return err
	}
	return nil
}

func renderTapTOML(c Config, items []Item) string {
	var b strings.Builder
	b.WriteString("# tap.toml — committed manifest of items in this tap.\n")
	b.WriteString("# Regenerate with `jwa-tobrew config`. Source of truth is the .rb files under Casks/ and Formula/.\n\n")
	fmt.Fprintf(&b, "[tap]\nowner = %q\nname  = %q\n\n", c.TapOwner, c.TapName)
	for _, it := range items {
		fmt.Fprintf(&b, "[[items]]\n")
		fmt.Fprintf(&b, "name    = %q\n", it.Name)
		fmt.Fprintf(&b, "kind    = %q\n", it.Kind)
		if it.Source != "" {
			fmt.Fprintf(&b, "source  = %q\n", it.Source)
		}
		fmt.Fprintf(&b, "version = %q\n", it.Version)
		if it.Asset != "" {
			fmt.Fprintf(&b, "asset   = %q\n", it.Asset)
		}
		if it.Desc != "" {
			fmt.Fprintf(&b, "desc    = %q\n", it.Desc)
		}
		if len(it.Deps) > 0 {
			fmt.Fprintf(&b, "deps    = [%s]\n", quoteList(it.Deps))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func quoteList(ss []string) string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(out, ", ")
}

func renderTapLocalTOML(items []Item, annots []LocalAnnotation) string {
	annByName := make(map[string]LocalAnnotation, len(annots))
	for _, a := range annots {
		annByName[a.Name] = a
	}
	var b strings.Builder
	b.WriteString("# tap.local.toml — gitignored, machine-specific.\n")
	b.WriteString("# Regenerate with `jwa-tobrew config`. Annotates each item with whether the\n")
	b.WriteString("# source repo is locally cloned (and where) — items without `local_clone`\n")
	b.WriteString("# exist as a snapshot of a published GitHub release only.\n\n")
	for _, it := range items {
		ann := annByName[it.Name]
		fmt.Fprintf(&b, "[[items]]\n")
		fmt.Fprintf(&b, "name        = %q\n", it.Name)
		if ann.LocalClone != "" {
			fmt.Fprintf(&b, "local_clone = %q\n", ann.LocalClone)
		} else {
			fmt.Fprintf(&b, "# local_clone = (none — snapshot from %s)\n", it.Source)
		}
		b.WriteString("\n")
	}
	return b.String()
}
