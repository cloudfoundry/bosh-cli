package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeDiskDeployer struct {
	DeployInputs []DiskDeployInput
	deployOutput diskDeployOutput
}

type DiskDeployInput struct {
	DiskPool         bmdepl.DiskPool
	Cloud            bmcloud.Cloud
	VM               bmvm.VM
	EventLoggerStage bmeventlog.Stage
}

type diskDeployOutput struct {
	err error
}

func NewFakeDiskDeployer() *FakeDiskDeployer {
	return &FakeDiskDeployer{
		DeployInputs: []DiskDeployInput{},
	}
}

func (d *FakeDiskDeployer) Deploy(
	diskPool bmdepl.DiskPool,
	cloud bmcloud.Cloud,
	vm bmvm.VM,
	eventLoggerStage bmeventlog.Stage,
) error {
	d.DeployInputs = append(d.DeployInputs, DiskDeployInput{
		DiskPool:         diskPool,
		Cloud:            cloud,
		VM:               vm,
		EventLoggerStage: eventLoggerStage,
	})

	return d.deployOutput.err
}

func (d *FakeDiskDeployer) SetDeployBehavior(err error) {
	d.deployOutput = diskDeployOutput{
		err: err,
	}
}
