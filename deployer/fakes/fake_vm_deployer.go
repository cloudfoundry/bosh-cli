package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeVMDeployer struct {
	DeployInputs  []VMDeployInput
	DeployOutputs []vmDeployOutput

	WaitUntilReadyInputs []WaitUntilReadyInput
	WaitUntilReadyErr    error
}

type VMDeployInput struct {
	Cloud            bmcloud.Cloud
	Deployment       bmdepl.Deployment
	Stemcell         bmstemcell.CloudStemcell
	MbusURL          string
	EventLoggerStage bmeventlog.Stage
}

type WaitUntilReadyInput struct {
	VM               bmvm.VM
	SSHTunnelOptions bmsshtunnel.Options
	EventLoggerStage bmeventlog.Stage
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
	deployment bmdepl.Deployment,
	stemcell bmstemcell.CloudStemcell,
	mbusURL string,
	eventLoggerStage bmeventlog.Stage,
) (bmvm.VM, error) {
	input := VMDeployInput{
		Cloud:            cloud,
		Deployment:       deployment,
		Stemcell:         stemcell,
		MbusURL:          mbusURL,
		EventLoggerStage: eventLoggerStage,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.vm, output.err
}

func (m *FakeVMDeployer) WaitUntilReady(vm bmvm.VM, sshTunnelOptions bmsshtunnel.Options, eventLoggerStage bmeventlog.Stage) error {
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
