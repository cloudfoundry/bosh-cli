package fakes

import (
	bmdeplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
)

type FakeParser struct {
	ParsePath     string
	ParseManifest bmdeplmanifest.Manifest
	ParseErr      error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (bmdeplmanifest.Manifest, error) {
	p.ParsePath = path
	return p.ParseManifest, p.ParseErr
}
