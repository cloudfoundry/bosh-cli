package fakes

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeParser struct {
	ParsePath          string
	ParseDeployment    bmdepl.Deployment
	ParseCPIDeployment bmdepl.CPIDeployment
	ParseErr           error
}

func NewFakeParser() *FakeParser {
	return &FakeParser{}
}

func (p *FakeParser) Parse(path string) (bmdepl.Deployment, bmdepl.CPIDeployment, error) {
	p.ParsePath = path
	return p.ParseDeployment, p.ParseCPIDeployment, p.ParseErr
}
