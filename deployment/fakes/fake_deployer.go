package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

type DeployInput struct {
	Cpi             bmcloud.Cloud
	Manifest        bmdeplmanifest.Manifest
	Stemcell        bmstemcell.ExtractedStemcell
	Registry        bminstallmanifest.Registry
	SSHTunnelConfig bminstallmanifest.SSHTunnel
	VMManager       bmvm.Manager
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
	deploymentManifest bmdeplmanifest.Manifest,
	stemcell bmstemcell.ExtractedStemcell,
	registry bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	vmManager bmvm.Manager,
) error {
	input := DeployInput{
		Cpi:             cpi,
		Manifest:        deploymentManifest,
		Stemcell:        stemcell,
		Registry:        registry,
		SSHTunnelConfig: sshTunnelConfig,
		VMManager:       vmManager,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.err
}

func (m *FakeDeployer) SetDeployBehavior(err error) {
	m.DeployOutputs = append(m.DeployOutputs, deployOutput{err: err})
}
