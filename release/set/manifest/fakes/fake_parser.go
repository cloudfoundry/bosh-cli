package fakes

import (
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
)

type FakeParser struct {
	ParsePath     string
	ParseManifest birelsetmanifest.Manifest
	ParseErr      error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string, vars boshtpl.Variables) (birelsetmanifest.Manifest, error) {
	p.ParsePath = path
	return p.ParseManifest, p.ParseErr
}
