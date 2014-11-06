package fakes

import (
	"time"

	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeInstance struct {
	ApplyInputs []ApplyInput
	ApplyErr    error

	StartCalled bool
	StartErr    error

	AttachDiskInputs []AttachDiskInput
	AttachDiskErr    error

	WaitToBeReadyInputs []WaitInput
	WaitToBeReadyErr    error

	WaitToBeRunningInputs []WaitInput
	WaitToBeRunningErr    error
}

type ApplyInput struct {
	StemcellApplySpec bmstemcell.ApplySpec
	Deployment        bmdepl.Deployment
}

type WaitInput struct {
	MaxAttempts int
	Delay       time.Duration
}

type AttachDiskInput struct {
	Disk bmdisk.Disk
}

func NewFakeInstance() *FakeInstance {
	return &FakeInstance{
		ApplyInputs:           []ApplyInput{},
		WaitToBeReadyInputs:   []WaitInput{},
		WaitToBeRunningInputs: []WaitInput{},
		AttachDiskInputs:      []AttachDiskInput{},
	}
}

func (i *FakeInstance) WaitToBeReady(maxAttempts int, delay time.Duration) error {
	i.WaitToBeReadyInputs = append(i.WaitToBeReadyInputs, WaitInput{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	})
	return i.WaitToBeReadyErr
}

func (i *FakeInstance) Apply(stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment) error {
	i.ApplyInputs = append(i.ApplyInputs, ApplyInput{
		StemcellApplySpec: stemcellApplySpec,
		Deployment:        deployment,
	})

	return i.ApplyErr
}

func (i *FakeInstance) Start() error {
	i.StartCalled = true

	return i.StartErr
}

func (i *FakeInstance) WaitToBeRunning(maxAttempts int, delay time.Duration) error {
	i.WaitToBeRunningInputs = append(i.WaitToBeRunningInputs, WaitInput{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	})
	return i.WaitToBeRunningErr
}

func (i *FakeInstance) AttachDisk(disk bmdisk.Disk) error {
	i.AttachDiskInputs = append(i.AttachDiskInputs, AttachDiskInput{
		Disk: disk,
	})

	return i.AttachDiskErr
}
