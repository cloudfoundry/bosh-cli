package fakes

import (
	"time"

	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeVM struct {
	cid string

	ExistsCalled int
	ExistsFound  bool
	ExistsErr    error

	UpdateDisksInputs []UpdateDisksInput
	UpdateDisksErr    error

	ApplyInputs []ApplyInput
	ApplyErr    error

	StartCalled int
	StartErr    error

	AttachDiskInputs   []AttachDiskInput
	attachDiskBehavior map[string]error

	DetachDiskInputs   []DetachDiskInput
	detachDiskBehavior map[string]error

	WaitUntilReadyInputs []WaitUntilReadyInput
	WaitUntilReadyErr    error

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

	MigrateDiskCalledTimes int
	MigrateDiskErr         error
}

type UpdateDisksInput struct {
	DiskPool bmmanifest.DiskPool
	Stage    bmeventlog.Stage
}

type ApplyInput struct {
	StemcellApplySpec bmstemcell.ApplySpec
	Manifest          bmmanifest.Manifest
}

type WaitUntilReadyInput struct {
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

type DetachDiskInput struct {
	Disk bmdisk.Disk
}

type UnmountDiskInput struct {
	Disk bmdisk.Disk
}

func NewFakeVM(cid string) *FakeVM {
	return &FakeVM{
		ExistsFound:           true,
		ApplyInputs:           []ApplyInput{},
		WaitUntilReadyInputs:  []WaitUntilReadyInput{},
		WaitToBeRunningInputs: []WaitInput{},
		AttachDiskInputs:      []AttachDiskInput{},
		DetachDiskInputs:      []DetachDiskInput{},
		UnmountDiskInputs:     []UnmountDiskInput{},
		attachDiskBehavior:    map[string]error{},
		detachDiskBehavior:    map[string]error{},
		cid:                   cid,
	}
}

func (vm *FakeVM) CID() string {
	return vm.cid
}

func (vm *FakeVM) Exists() (bool, error) {
	vm.ExistsCalled++
	return vm.ExistsFound, vm.ExistsErr
}

func (vm *FakeVM) WaitUntilReady(timeout time.Duration, delay time.Duration) error {
	vm.WaitUntilReadyInputs = append(vm.WaitUntilReadyInputs, WaitUntilReadyInput{
		Timeout: timeout,
		Delay:   delay,
	})
	return vm.WaitUntilReadyErr
}

func (vm *FakeVM) UpdateDisks(diskPool bmmanifest.DiskPool, eventLoggerStage bmeventlog.Stage) error {
	vm.UpdateDisksInputs = append(vm.UpdateDisksInputs, UpdateDisksInput{
		DiskPool: diskPool,
		Stage:    eventLoggerStage,
	})
	return vm.UpdateDisksErr
}

func (vm *FakeVM) Apply(stemcellApplySpec bmstemcell.ApplySpec, deploymentManifest bmmanifest.Manifest) error {
	vm.ApplyInputs = append(vm.ApplyInputs, ApplyInput{
		StemcellApplySpec: stemcellApplySpec,
		Manifest:          deploymentManifest,
	})

	return vm.ApplyErr
}

func (vm *FakeVM) Start() error {
	vm.StartCalled++
	return vm.StartErr
}

func (vm *FakeVM) WaitToBeRunning(maxAttempts int, delay time.Duration) error {
	vm.WaitToBeRunningInputs = append(vm.WaitToBeRunningInputs, WaitInput{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	})
	return vm.WaitToBeRunningErr
}

func (vm *FakeVM) AttachDisk(disk bmdisk.Disk) error {
	vm.AttachDiskInputs = append(vm.AttachDiskInputs, AttachDiskInput{
		Disk: disk,
	})

	return vm.attachDiskBehavior[disk.CID()]
}

func (vm *FakeVM) DetachDisk(disk bmdisk.Disk) error {
	vm.DetachDiskInputs = append(vm.DetachDiskInputs, DetachDiskInput{
		Disk: disk,
	})

	return vm.detachDiskBehavior[disk.CID()]
}

func (vm *FakeVM) UnmountDisk(disk bmdisk.Disk) error {
	vm.UnmountDiskInputs = append(vm.UnmountDiskInputs, UnmountDiskInput{
		Disk: disk,
	})

	return vm.UnmountDiskErr
}

func (vm *FakeVM) MigrateDisk() error {
	vm.MigrateDiskCalledTimes++

	return vm.MigrateDiskErr
}

func (vm *FakeVM) Stop() error {
	vm.StopCalled++
	return vm.StopErr
}

func (vm *FakeVM) Disks() ([]bmdisk.Disk, error) {
	return vm.ListDisksDisks, vm.ListDisksErr
}

func (vm *FakeVM) Delete() error {
	vm.DeleteCalled++
	return vm.DeleteErr
}

func (vm *FakeVM) SetAttachDiskBehavior(disk bmdisk.Disk, err error) {
	vm.attachDiskBehavior[disk.CID()] = err
}

func (vm *FakeVM) SetDetachDiskBehavior(disk bmdisk.Disk, err error) {
	vm.detachDiskBehavior[disk.CID()] = err
}
