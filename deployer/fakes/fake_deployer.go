package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type DeployInput struct {
	Cpi               bmcloud.Cloud
	Deployment        bmdepl.Deployment
	StemcellApplySpec bmstemcell.ApplySpec
	Registry          bmdepl.Registry
	SSHTunnelConfig   bmdepl.SSHTunnel
	MbusURL           string
	Stemcell          bmstemcell.CloudStemcell
}

type deployOutput struct {
	err error
}

type FakeDeployer struct {
	DeployInputs  []DeployInput
	DeployOutputs []deployOutput
}

func NewFakeDeployer() *FakeDeployer {
	return &FakeDeployer{
		DeployInputs:  []DeployInput{},
		DeployOutputs: []deployOutput{},
	}
}

func (m *FakeDeployer) Deploy(
	cpi bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcellApplySpec bmstemcell.ApplySpec,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	mbusURL string,
	stemcell bmstemcell.CloudStemcell,
) error {
	input := DeployInput{
		Cpi:               cpi,
		Deployment:        deployment,
		StemcellApplySpec: stemcellApplySpec,
		Registry:          registry,
		SSHTunnelConfig:   sshTunnelConfig,
		MbusURL:           mbusURL,
		Stemcell:          stemcell,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.err
}

func (m *FakeDeployer) SetDeployBehavior(err error) {
	m.DeployOutputs = append(m.DeployOutputs, deployOutput{err: err})
}
