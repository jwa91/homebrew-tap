package main

import _ "embed"

//go:embed templates/goreleaser.yaml.tmpl
var goreleaserTmpl string

// Used for cask + formula scaffolds. Go projects use .goreleaser.yaml
// directly (no shell wrapper) so they have a unified release pipeline
// across local + CI runs.
//
//go:embed templates/release.sh.tmpl
var releaseShTmpl string

//go:embed templates/cask.rb.tmpl
var caskRbTmpl string

//go:embed templates/formula.rb.tmpl
var formulaRbTmpl string

//go:embed templates/skill-release.md.tmpl
var skillReleaseTmpl string
