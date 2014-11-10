package fakes

import (
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bmdisk.Disk
	CreateErr    error

	FindOutput FindOutput
}

type CreateInput struct {
	DiskPool   bmdepl.DiskPool
	InstanceID string
}

type FindOutput struct {
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

func (m *FakeManager) Find() (bmdisk.Disk, bool, error) {
	return m.FindOutput.Disk, m.FindOutput.Found, m.FindOutput.Err
}

func (m *FakeManager) SetFindBehavior(disk bmdisk.Disk, found bool, err error) {
	m.FindOutput = FindOutput{
		Disk:  disk,
		Found: found,
		Err:   err,
	}
}
