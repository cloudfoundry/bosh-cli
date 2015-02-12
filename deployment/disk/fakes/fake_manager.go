package fakes

import (
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bmdisk.Disk
	CreateErr    error

	findCurrentOutput findCurrentOutput

	DeleteUnusedCalledTimes int
	DeleteUnusedErr         error

	findUnusedOutput findUnusedOutput
}

type CreateInput struct {
	DiskPool   bmdeplmanifest.DiskPool
	InstanceID string
}

type findCurrentOutput struct {
	Disks []bmdisk.Disk
	Err   error
}

type findUnusedOutput struct {
	disks []bmdisk.Disk
	err   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(diskPool bmdeplmanifest.DiskPool, instanceID string) (bmdisk.Disk, error) {
	input := CreateInput{
		DiskPool:   diskPool,
		InstanceID: instanceID,
	}
	m.CreateInputs = append(m.CreateInputs, input)

	return m.CreateDisk, m.CreateErr
}

func (m *FakeManager) FindCurrent() ([]bmdisk.Disk, error) {
	return m.findCurrentOutput.Disks, m.findCurrentOutput.Err
}

func (m *FakeManager) FindUnused() ([]bmdisk.Disk, error) {
	return m.findUnusedOutput.disks, m.findUnusedOutput.err
}

func (m *FakeManager) DeleteUnused(eventLogStage bmui.Stage) error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
}

func (m *FakeManager) SetFindCurrentBehavior(disks []bmdisk.Disk, err error) {
	m.findCurrentOutput = findCurrentOutput{
		Disks: disks,
		Err:   err,
	}
}

func (m *FakeManager) SetFindUnusedBehavior(
	disks []bmdisk.Disk,
	err error,
) {
	m.findUnusedOutput = findUnusedOutput{
		disks: disks,
		err:   err,
	}
}
