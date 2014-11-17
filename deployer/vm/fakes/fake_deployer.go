package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type DeployInput struct {
	Cloud            bmcloud.Cloud
	Deployment       bmdepl.Deployment
	Stemcell         bmstemcell.CloudStemcell
	SSHTunnelOptions bmsshtunnel.Options
	MbusURL          string
	EventLoggerStage bmeventlog.Stage
}

type deployOutput struct {
	vm  bmvm.VM
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
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcell bmstemcell.CloudStemcell,
	sshTunnelOptions bmsshtunnel.Options,
	mbusURL string,
	eventLoggerStage bmeventlog.Stage,
) (bmvm.VM, error) {
	input := DeployInput{
		Cloud:            cloud,
		Deployment:       deployment,
		Stemcell:         stemcell,
		SSHTunnelOptions: sshTunnelOptions,
		MbusURL:          mbusURL,
		EventLoggerStage: eventLoggerStage,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.vm, output.err
}

func (m *FakeDeployer) SetDeployBehavior(vm bmvm.VM, err error) {
	m.DeployOutputs = append(m.DeployOutputs, deployOutput{vm: vm, err: err})
}
