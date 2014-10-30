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

type FakeMicroDeployer struct {
	DeployInput  DeployInput
	DeployOutput deployOutput
}

func NewFakeMicroDeployer() *FakeMicroDeployer {
	return &FakeMicroDeployer{
		DeployInput: DeployInput{},
	}
}

func (m *FakeMicroDeployer) Deploy(
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

func (m *FakeMicroDeployer) SetDeployBehavior(err error) {
	m.DeployOutput = deployOutput{err: err}
}
