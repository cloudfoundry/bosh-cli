package fakes

import (
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bmdisk.Disk
	CreateErr    error

	findCurrentOutput findCurrentOutput

	DeleteUnusedCalledTimes int
	DeleteUnusedErr         error
}

type CreateInput struct {
	DiskPool   bmdepl.DiskPool
	InstanceID string
}

type findCurrentOutput struct {
	Disk  bmdisk.Disk
	Found bool
	Err   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(diskPool bmdepl.DiskPool, instanceID string) (bmdisk.Disk, error) {
	input := CreateInput{
		DiskPool:   diskPool,
		InstanceID: instanceID,
	}
	m.CreateInputs = append(m.CreateInputs, input)

	return m.CreateDisk, m.CreateErr
}

func (m *FakeManager) FindCurrent() (bmdisk.Disk, bool, error) {
	return m.findCurrentOutput.Disk, m.findCurrentOutput.Found, m.findCurrentOutput.Err
}

func (m *FakeManager) DeleteUnused() error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
}

func (m *FakeManager) SetFindCurrentBehavior(disk bmdisk.Disk, found bool, err error) {
	m.findCurrentOutput = findCurrentOutput{
		Disk:  disk,
		Found: found,
		Err:   err,
	}
}
