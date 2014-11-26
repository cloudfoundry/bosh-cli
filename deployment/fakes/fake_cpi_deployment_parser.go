package fakes

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeCPIDeploymentParser struct {
	ParsePath       string
	ParseDeployment bmdepl.CPIDeployment
	ParseErr        error
}

func NewFakeCPIDeploymentParser() *FakeCPIDeploymentParser {
	return &FakeCPIDeploymentParser{}
}

func (p *FakeCPIDeploymentParser) Parse(path string) (bmdepl.CPIDeployment, error) {
	p.ParsePath = path
	return p.ParseDeployment, p.ParseErr
}
