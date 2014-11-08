package fakes

import (
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type CreateInput struct {
	DiskPool   bmdepl.DiskPool
	InstanceID string
}

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bmdisk.Disk
	CreateErr    error
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
