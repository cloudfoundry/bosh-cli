package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type DeployInput struct {
	StemcellCID bmstemcell.CID
	Cloud       bmcloud.Cloud
	Deployment  bmdepl.Deployment
}

type deployOutput struct {
	err error
}

type FakeMicroDeployer struct {
	DeployInput  DeployInput
	DeployOutput deployOutput
}

func NewFakeMicroDeployer() *FakeMicroDeployer {
	return &FakeMicroDeployer{
		DeployInput: DeployInput{},
	}
}

func (m *FakeMicroDeployer) Deploy(cloud bmcloud.Cloud, deployment bmdepl.Deployment, stemcellCID bmstemcell.CID) error {
	input := DeployInput{
		StemcellCID: stemcellCID,
		Cloud:       cloud,
		Deployment:  deployment,
	}
	m.DeployInput = input

	if (m.DeployOutput != deployOutput{}) {
		return m.DeployOutput.err
	}

	return nil
}

func (m *FakeMicroDeployer) SetDeployBehavior(err error) {
	m.DeployOutput = deployOutput{err: err}
}
