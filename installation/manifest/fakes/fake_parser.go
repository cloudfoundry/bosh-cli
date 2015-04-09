package fakes

import (
	bminstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
)

type FakeParser struct {
	ParsePath     string
	ParseManifest bminstallmanifest.Manifest
	ParseErr      error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (bminstallmanifest.Manifest, error) {
	p.ParsePath = path
	return p.ParseManifest, p.ParseErr
}
