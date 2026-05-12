package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runUpgrade(args []string) error {
	_ = subFlagSet("upgrade", "re-install jwa-tobrew via brew").Parse(args)
	c := LoadConfig()
	target := fmt.Sprintf("%s/%s/jwa-tobrew", c.TapOwner, c.TapName)
	info("brew update")
	if err := runIO("brew", "update"); err != nil {
		return err
	}
	info("brew upgrade %s", target)
	return runIO("brew", "upgrade", target)
}

// runIO runs a command wired to the parent's stdio.
func runIO(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
