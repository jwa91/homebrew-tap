package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// isGitRepo returns true if dir is inside a git work tree.
func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// repoOriginInfo parses owner/repo out of `origin`'s URL.
func repoOriginInfo(dir string) (string, string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", "", errors.New("no `origin` remote configured")
	}
	return parseRepoRef(strings.TrimSpace(string(out)))
}

func gitDirty(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// hasHeadTag returns true if the current HEAD has any annotated/lightweight tag.
func hasHeadTag(dir string) bool {
	cmd := exec.Command("git", "describe", "--exact-match", "--tags", "HEAD")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// tagAndPush creates an annotated tag at HEAD (if missing) and pushes it.
func tagAndPush(dir, tag string, push bool) error {
	check := exec.Command("git", "rev-parse", "-q", "--verify", "refs/tags/"+tag)
	check.Dir = dir
	if check.Run() != nil {
		info("creating tag %s", tag)
		create := exec.Command("git", "tag", "-a", tag, "-m", tag)
		create.Dir = dir
		if out, err := create.CombinedOutput(); err != nil {
			return fmt.Errorf("git tag %s: %s", tag, strings.TrimSpace(string(out)))
		}
	} else {
		info("tag %s already exists", tag)
	}
	if !push {
		return nil
	}
	info("pushing tag %s", tag)
	p := exec.Command("git", "push", "origin", tag)
	p.Dir = dir
	if out, err := p.CombinedOutput(); err != nil {
		return fmt.Errorf("git push origin %s: %s", tag, strings.TrimSpace(string(out)))
	}
	return nil
}
