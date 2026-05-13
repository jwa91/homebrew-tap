package main

import "testing"

// TidyDep locks in the deps-serialization fix: a bare quoted formula
// dependency (e.g. `depends_on "git"`) must become `git` so the manifest
// renderer doesn't re-wrap it as `"\"git\""`. Cask-style key:value deps
// keep their inner quotes.
func TestTidyDep(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`"git"`, "git"},
		{`  "go"  `, "go"},
		{`"git" # comment`, "git"},
		{`macos: ">= :tahoe"`, `macos: ">= :tahoe"`},
		{`arch: :arm64`, "arch: :arm64"},
		{`"git" => :build`, `"git" => :build`},
	}
	for _, tc := range cases {
		got := tidyDep(tc.in)
		if got != tc.want {
			t.Errorf("tidyDep(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
