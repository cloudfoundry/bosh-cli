package fakes

import (
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

type FakeParser struct {
	ParsePath                  string
	ParseDeployment            bmmanifest.Manifest // todo
	ParseCPIDeploymentManifest bmmanifest.CPIDeploymentManifest
	ParseErr                   error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (bmmanifest.Manifest, bmmanifest.CPIDeploymentManifest, error) {
	p.ParsePath = path
	return p.ParseDeployment, p.ParseCPIDeploymentManifest, p.ParseErr
}
