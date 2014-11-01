package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type DeployInput struct {
	Cpi               bmcloud.Cloud
	Deployment        bmdepl.Deployment
	StemcellApplySpec bmstemcell.ApplySpec
	Registry          bmdepl.Registry
	SSHTunnelConfig   bmdepl.SSHTunnel
	MbusURL           string
	StemcellCID       bmstemcell.CID
}

type deployOutput struct {
	err error
}

type FakeDeployer struct {
	DeployInput  DeployInput
	DeployOutput deployOutput
}

func NewFakeDeployer() *FakeDeployer {
	return &FakeDeployer{
		DeployInput: DeployInput{},
	}
}

func (m *FakeDeployer) Deploy(
	cpi bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcellApplySpec bmstemcell.ApplySpec,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	mbusURL string,
	stemcellCID bmstemcell.CID,
) error {
	input := DeployInput{
		Cpi:               cpi,
		Deployment:        deployment,
		StemcellApplySpec: stemcellApplySpec,
		Registry:          registry,
		SSHTunnelConfig:   sshTunnelConfig,
		MbusURL:           mbusURL,
		StemcellCID:       stemcellCID,
	}
	m.DeployInput = input

	if (m.DeployOutput != deployOutput{}) {
		return m.DeployOutput.err
	}

	return nil
}

func (m *FakeDeployer) SetDeployBehavior(err error) {
	m.DeployOutput = deployOutput{err: err}
}
