package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type FakeVMDeployer struct {
	DeployInputs  []VMDeployInput
	DeployOutputs []vmDeployOutput

	WaitUntilReadyInputs []WaitUntilReadyInput
	WaitUntilReadyErr    error
}

type VMDeployInput struct {
	Cloud            bmcloud.Cloud
	Manifest         bmdeplmanifest.Manifest
	Stemcell         bmstemcell.CloudStemcell
	MbusURL          string
	EventLoggerStage bmui.Stage
}

type WaitUntilReadyInput struct {
	VM               bmvm.VM
	SSHTunnelOptions bmsshtunnel.Options
	EventLoggerStage bmui.Stage
}

type vmDeployOutput struct {
	vm  bmvm.VM
	err error
}

func NewFakeVMDeployer() *FakeVMDeployer {
	return &FakeVMDeployer{
		DeployInputs:  []VMDeployInput{},
		DeployOutputs: []vmDeployOutput{},
	}
}

func (m *FakeVMDeployer) Deploy(
	cloud bmcloud.Cloud,
	deploymentManifest bmdeplmanifest.Manifest,
	stemcell bmstemcell.CloudStemcell,
	mbusURL string,
	eventLoggerStage bmui.Stage,
) (bmvm.VM, error) {
	input := VMDeployInput{
		Cloud:            cloud,
		Manifest:         deploymentManifest,
		Stemcell:         stemcell,
		MbusURL:          mbusURL,
		EventLoggerStage: eventLoggerStage,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.vm, output.err
}

func (m *FakeVMDeployer) WaitUntilReady(vm bmvm.VM, sshTunnelOptions bmsshtunnel.Options, eventLoggerStage bmui.Stage) error {
	input := WaitUntilReadyInput{
		VM:               vm,
		SSHTunnelOptions: sshTunnelOptions,
		EventLoggerStage: eventLoggerStage,
	}
	m.WaitUntilReadyInputs = append(m.WaitUntilReadyInputs, input)

	return m.WaitUntilReadyErr
}

func (m *FakeVMDeployer) SetDeployBehavior(vm bmvm.VM, err error) {
	m.DeployOutputs = append(m.DeployOutputs, vmDeployOutput{vm: vm, err: err})
}
