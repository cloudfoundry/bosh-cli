package fakes

import (
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
)

type FakeParser struct {
	ParsePath     string
	ParseManifest bideplmanifest.Manifest
	ParseErr      error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string, vars boshtpl.Variables) (bideplmanifest.Manifest, error) {
	p.ParsePath = path
	return p.ParseManifest, p.ParseErr
}
