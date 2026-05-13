package main

import _ "embed"

//go:embed templates/goreleaser.yaml.tmpl
var goreleaserTmpl string

//go:embed templates/release-go.sh.tmpl
var releaseShTmpl string

//go:embed templates/release-cask.sh.tmpl
var releaseShCaskTmpl string

//go:embed templates/cask.rb.tmpl
var caskRbTmpl string

//go:embed templates/formula.rb.tmpl
var formulaRbTmpl string

//go:embed templates/skill-release.md.tmpl
var skillReleaseTmpl string
