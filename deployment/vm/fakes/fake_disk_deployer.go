package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeDiskDeployer struct {
	DeployInputs  []DeployInput
	deployOutputs deployOutput
}

type DeployInput struct {
	DiskPool         bmmanifest.DiskPool
	Cloud            bmcloud.Cloud
	VM               bmvm.VM
	EventLoggerStage bmeventlog.Stage
}

type deployOutput struct {
	err error
}

func NewFakeDiskDeployer() *FakeDiskDeployer {
	return &FakeDiskDeployer{
		DeployInputs: []DeployInput{},
	}
}

func (d *FakeDiskDeployer) Deploy(
	diskPool bmmanifest.DiskPool,
	cloud bmcloud.Cloud,
	vm bmvm.VM,
	eventLoggerStage bmeventlog.Stage,
) error {
	d.DeployInputs = append(d.DeployInputs, DeployInput{
		DiskPool:         diskPool,
		Cloud:            cloud,
		VM:               vm,
		EventLoggerStage: eventLoggerStage,
	})

	return d.deployOutputs.err
}

func (d *FakeDiskDeployer) SetDeployBehavior(err error) {
	d.deployOutputs = deployOutput{
		err: err,
	}
}
