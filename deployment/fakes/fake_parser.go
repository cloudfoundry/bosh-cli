package fakes

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeParser struct {
	ParsePath                  string
	ParseDeployment            bmdepl.Deployment
	ParseCPIDeploymentManifest bmdepl.CPIDeploymentManifest
	ParseErr                   error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (bmdepl.Deployment, bmdepl.CPIDeploymentManifest, error) {
	p.ParsePath = path
	return p.ParseDeployment, p.ParseCPIDeploymentManifest, p.ParseErr
}
