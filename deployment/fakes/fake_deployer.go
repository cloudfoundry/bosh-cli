package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
)

type DeployInput struct {
	Cpi             bmcloud.Cloud
	Manifest        bmmanifest.Manifest
	Stemcell        bmstemcell.ExtractedStemcell
	Registry        bmmanifest.Registry
	SSHTunnelConfig bmmanifest.SSHTunnel
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
	deploymentManifest bmmanifest.Manifest,
	stemcell bmstemcell.ExtractedStemcell,
	registry bmmanifest.Registry,
	sshTunnelConfig bmmanifest.SSHTunnel,
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
