package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type FakeDiskDeployer struct {
	DeployInputs  []DeployInput
	deployOutputs deployOutput
}

type DeployInput struct {
	DiskPool         bmdeplmanifest.DiskPool
	Cloud            bmcloud.Cloud
	VM               bmvm.VM
	EventLoggerStage bmui.Stage
}

type deployOutput struct {
	disks []bmdisk.Disk
	err   error
}

func NewFakeDiskDeployer() *FakeDiskDeployer {
	return &FakeDiskDeployer{
		DeployInputs: []DeployInput{},
	}
}

func (d *FakeDiskDeployer) Deploy(
	diskPool bmdeplmanifest.DiskPool,
	cloud bmcloud.Cloud,
	vm bmvm.VM,
	eventLoggerStage bmui.Stage,
) ([]bmdisk.Disk, error) {
	d.DeployInputs = append(d.DeployInputs, DeployInput{
		DiskPool:         diskPool,
		Cloud:            cloud,
		VM:               vm,
		EventLoggerStage: eventLoggerStage,
	})

	return d.deployOutputs.disks, d.deployOutputs.err
}

func (d *FakeDiskDeployer) SetDeployBehavior(disks []bmdisk.Disk, err error) {
	d.deployOutputs = deployOutput{
		disks: disks,
		err:   err,
	}
}
