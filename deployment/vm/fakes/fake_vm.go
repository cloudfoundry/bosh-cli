package fakes

import (
	"time"

	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type FakeVM struct {
	cid string

	ExistsCalled int
	ExistsFound  bool
	ExistsErr    error

	AgentClientReturn bmagentclient.AgentClient

	UpdateDisksInputs []UpdateDisksInput
	UpdateDisksDisks  []bmdisk.Disk
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
	DiskPool bmdeplmanifest.DiskPool
	Stage    bmui.Stage
}

type ApplyInput struct {
	ApplySpec bmas.ApplySpec
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

func (vm *FakeVM) AgentClient() bmagentclient.AgentClient {
	return vm.AgentClientReturn
}

func (vm *FakeVM) WaitUntilReady(timeout time.Duration, delay time.Duration) error {
	vm.WaitUntilReadyInputs = append(vm.WaitUntilReadyInputs, WaitUntilReadyInput{
		Timeout: timeout,
		Delay:   delay,
	})
	return vm.WaitUntilReadyErr
}

func (vm *FakeVM) UpdateDisks(diskPool bmdeplmanifest.DiskPool, eventLoggerStage bmui.Stage) ([]bmdisk.Disk, error) {
	vm.UpdateDisksInputs = append(vm.UpdateDisksInputs, UpdateDisksInput{
		DiskPool: diskPool,
		Stage:    eventLoggerStage,
	})
	return vm.UpdateDisksDisks, vm.UpdateDisksErr
}

func (vm *FakeVM) Apply(applySpec bmas.ApplySpec) error {
	vm.ApplyInputs = append(vm.ApplyInputs, ApplyInput{
		ApplySpec: applySpec,
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
