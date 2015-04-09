package fakes

import (
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
)

type FakeParser struct {
	ParsePath     string
	ParseManifest biinstallmanifest.Manifest
	ParseErr      error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (biinstallmanifest.Manifest, error) {
	p.ParsePath = path
	return p.ParseManifest, p.ParseErr
}
