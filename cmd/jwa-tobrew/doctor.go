package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runDoctor(args []string) error {
	fs := subFlagSet("doctor", "check tools, tap location, SSH origin, and required env")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	banner("jwa-tobrew doctor")
	c := LoadConfig()
	failed := 0

	for _, b := range []struct{ bin, brewPkg string }{
		{"git", "git"},
		{"gh", "gh"},
	} {
		if _, err := exec.LookPath(b.bin); err != nil {
			fail("%s not found — `brew install %s`", b.bin, b.brewPkg)
			failed++
		} else {
			ok("%s installed", b.bin)
		}
	}
	if _, err := exec.LookPath("goreleaser"); err != nil {
		warn("goreleaser not found — required for Go projects (`brew install goreleaser`)")
	} else {
		ok("goreleaser installed")
	}

	tap, err := TapDir(c)
	if err != nil {
		fail("%v", err)
		failed++
	} else {
		ok("tap repo at %s", tap)
		if err := requireSSHOrigin(tap); err != nil {
			fail("%v", err)
			failed++
		} else {
			ok("tap origin is SSH")
		}
	}

	// Token presence — validation and rotation are harden's job.
	if os.Getenv("GITHUB_TOKEN") == "" && os.Getenv("GH_TOKEN") == "" {
		warn("$GITHUB_TOKEN / $GH_TOKEN not set — release flows that hit the GH API will fail")
		hint("wrap with: jwa-harden run -- jwa-tobrew <command>")
	} else {
		ok("$GITHUB_TOKEN present")
	}

	if failed > 0 {
		return fmt.Errorf("%d check(s) failed", failed)
	}
	return nil
}

// requireGitHubToken is called from any subcommand that hits the GitHub API.
func requireGitHubToken() error {
	if os.Getenv("GITHUB_TOKEN") != "" || os.Getenv("GH_TOKEN") != "" {
		return nil
	}
	return fmt.Errorf("$GITHUB_TOKEN not set — wrap with: jwa-harden run -- jwa-tobrew <command>")
}
