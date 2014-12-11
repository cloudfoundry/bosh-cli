package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type DeployInput struct {
	Cpi             bmcloud.Cloud
	Manifest        bmdepl.Manifest
	Stemcell        bmstemcell.ExtractedStemcell
	Registry        bmdepl.Registry
	SSHTunnelConfig bmdepl.SSHTunnel
	MbusURL         string
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
	deploymentManifest bmdepl.Manifest,
	stemcell bmstemcell.ExtractedStemcell,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	mbusURL string,
) error {
	input := DeployInput{
		Cpi:             cpi,
		Manifest:        deploymentManifest,
		Stemcell:        stemcell,
		Registry:        registry,
		SSHTunnelConfig: sshTunnelConfig,
		MbusURL:         mbusURL,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.err
}

func (m *FakeDeployer) SetDeployBehavior(err error) {
	m.DeployOutputs = append(m.DeployOutputs, deployOutput{err: err})
}
