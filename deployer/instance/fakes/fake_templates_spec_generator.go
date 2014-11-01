package fakes

import (
	bmins "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type FakeTemplatesSpecGenerator struct {
	CreateInputs        []CreateInput
	CreateTemplatesSpec bmins.TemplatesSpec
	CreateErr           error
	CreateCalled        bool
}

type CreateInput struct {
	DeploymentJob  bmdepl.Job
	StemcellJob    bmstemcell.Job
	DeploymentName string
	Properties     map[string]interface{}
	MbusURL        string
}

func NewFakeTemplatesSpecGenerator() *FakeTemplatesSpecGenerator {
	return &FakeTemplatesSpecGenerator{
		CreateInputs: []CreateInput{},
	}
}

func (g *FakeTemplatesSpecGenerator) Create(
	deploymentJob bmdepl.Job,
	stemcellJob bmstemcell.Job,
	deploymentName string,
	properties map[string]interface{},
	mbusURL string,
) (bmins.TemplatesSpec, error) {
	g.CreateInputs = append(g.CreateInputs, CreateInput{
		DeploymentJob:  deploymentJob,
		StemcellJob:    stemcellJob,
		DeploymentName: deploymentName,
		Properties:     properties,
		MbusURL:        mbusURL,
	})

	g.CreateCalled = true
	return g.CreateTemplatesSpec, g.CreateErr
}

func (g *FakeTemplatesSpecGenerator) SetCreateBehavior(createTemplatesSpec bmins.TemplatesSpec, err error) {
	g.CreateTemplatesSpec = createTemplatesSpec
	g.CreateErr = err
}
