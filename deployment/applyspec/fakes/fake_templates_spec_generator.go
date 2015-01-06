package fakes

import (
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type FakeTemplatesSpecGenerator struct {
	CreateTemplatesSpecInputs []CreateTemplatesSpecInput
	CreateTemplatesSpec       bmas.TemplatesSpec
	CreateErr                 error
	CreateCalled              bool
}

type CreateTemplatesSpecInput struct {
	DeploymentJob  bmdeplmanifest.Job
	StemcellJob    bmstemcell.Job
	DeploymentName string
	Properties     map[string]interface{}
	MbusURL        string
}

func NewFakeTemplatesSpecGenerator() *FakeTemplatesSpecGenerator {
	return &FakeTemplatesSpecGenerator{
		CreateTemplatesSpecInputs: []CreateTemplatesSpecInput{},
	}
}

func (g *FakeTemplatesSpecGenerator) Create(
	deploymentJob bmdeplmanifest.Job,
	stemcellJob bmstemcell.Job,
	deploymentName string,
	properties map[string]interface{},
	mbusURL string,
) (bmas.TemplatesSpec, error) {
	g.CreateTemplatesSpecInputs = append(g.CreateTemplatesSpecInputs, CreateTemplatesSpecInput{
		DeploymentJob:  deploymentJob,
		StemcellJob:    stemcellJob,
		DeploymentName: deploymentName,
		Properties:     properties,
		MbusURL:        mbusURL,
	})

	g.CreateCalled = true
	return g.CreateTemplatesSpec, g.CreateErr
}

func (g *FakeTemplatesSpecGenerator) SetCreateBehavior(createTemplatesSpec bmas.TemplatesSpec, err error) {
	g.CreateTemplatesSpec = createTemplatesSpec
	g.CreateErr = err
}
