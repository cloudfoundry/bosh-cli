package fakes

import (
	"time"

	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeVM struct {
	cid string

	ApplyInputs []ApplyInput
	ApplyErr    error

	StartCalled int
	StartErr    error

	AttachDiskInputs []AttachDiskInput
	AttachDiskErr    error

	WaitToBeReadyInputs []WaitToBeReadyInput
	WaitToBeReadyErr    error

	WaitToBeRunningInputs []WaitInput
	WaitToBeRunningErr    error

	DeleteCalled int
	DeleteErr    error

	StopCalled int
	StopErr    error

	ListDisksDisks []bmdisk.Disk
	ListDisksErr   error

	UnmountDiskInputs []UnmountDiskInput
	UnmountDiskErr    error
}

type ApplyInput struct {
	StemcellApplySpec bmstemcell.ApplySpec
	Deployment        bmdepl.Deployment
}

type WaitToBeReadyInput struct {
	Timeout time.Duration
	Delay   time.Duration
}

type WaitInput struct {
	MaxAttempts int
	Delay       time.Duration
}

type AttachDiskInput struct {
	Disk bmdisk.Disk
}

type UnmountDiskInput struct {
	Disk bmdisk.Disk
}

func NewFakeVM(cid string) *FakeVM {
	return &FakeVM{
		ApplyInputs:           []ApplyInput{},
		WaitToBeReadyInputs:   []WaitToBeReadyInput{},
		WaitToBeRunningInputs: []WaitInput{},
		AttachDiskInputs:      []AttachDiskInput{},
		UnmountDiskInputs:     []UnmountDiskInput{},
		cid:                   cid,
	}
}

func (i *FakeVM) CID() string {
	return i.cid
}

func (i *FakeVM) WaitToBeReady(timeout time.Duration, delay time.Duration) error {
	i.WaitToBeReadyInputs = append(i.WaitToBeReadyInputs, WaitToBeReadyInput{
		Timeout: timeout,
		Delay:   delay,
	})
	return i.WaitToBeReadyErr
}

func (i *FakeVM) Apply(stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment) error {
	i.ApplyInputs = append(i.ApplyInputs, ApplyInput{
		StemcellApplySpec: stemcellApplySpec,
		Deployment:        deployment,
	})

	return i.ApplyErr
}

func (i *FakeVM) Start() error {
	i.StartCalled++
	return i.StartErr
}

func (i *FakeVM) WaitToBeRunning(maxAttempts int, delay time.Duration) error {
	i.WaitToBeRunningInputs = append(i.WaitToBeRunningInputs, WaitInput{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	})
	return i.WaitToBeRunningErr
}

func (i *FakeVM) AttachDisk(disk bmdisk.Disk) error {
	i.AttachDiskInputs = append(i.AttachDiskInputs, AttachDiskInput{
		Disk: disk,
	})

	return i.AttachDiskErr
}

func (i *FakeVM) UnmountDisk(disk bmdisk.Disk) error {
	i.UnmountDiskInputs = append(i.UnmountDiskInputs, UnmountDiskInput{
		Disk: disk,
	})

	return i.UnmountDiskErr
}

func (i *FakeVM) Stop() error {
	i.StopCalled++
	return i.StopErr
}

func (i *FakeVM) Disks() ([]bmdisk.Disk, error) {
	return i.ListDisksDisks, i.ListDisksErr
}

func (i *FakeVM) Delete() error {
	i.DeleteCalled++
	return i.DeleteErr
}
