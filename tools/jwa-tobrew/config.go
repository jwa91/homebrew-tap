package main

import "os"

// Config holds runtime configuration resolved from env vars.
// Per the security ground rules, no secrets are baked into the source.
// Token plumbing is owned by `jwa-harden run` — `jwa-tobrew` only consumes
// whatever env it inherits.
type Config struct {
	TapDir   string // BREWTAP_DIR (auto-detected if empty)
	TapOwner string // BREWTAP_TAP_OWNER (default "jwa91")
	TapName  string // BREWTAP_TAP_NAME (default "tap")
}

func LoadConfig() Config {
	return Config{
		TapDir:   os.Getenv("BREWTAP_DIR"),
		TapOwner: envOr("BREWTAP_TAP_OWNER", "jwa91"),
		TapName:  envOr("BREWTAP_TAP_NAME", "tap"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// runConfig regenerates tap.toml + tap.local.toml from the current state of
// the tap. Read-only with respect to .rb files; writes only the manifests.
func runConfig(args []string) error {
	fs := subFlagSet("config", "regenerate tap.toml + tap.local.toml from current tap state")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	c := LoadConfig()
	tap, err := TapDir(c)
	if err != nil {
		return err
	}
	if err := writeManifests(c, tap); err != nil {
		return err
	}
	ok("wrote tap.toml")
	ok("wrote tap.local.toml")
	ok("updated README.md items section")
	return nil
}
