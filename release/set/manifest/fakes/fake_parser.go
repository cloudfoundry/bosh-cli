package fakes

import (
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"
)

type FakeParser struct {
	ParsePath     string
	ParseManifest bmrelsetmanifest.Manifest
	ParseErr      error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (bmrelsetmanifest.Manifest, error) {
	p.ParsePath = path
	return p.ParseManifest, p.ParseErr
}
