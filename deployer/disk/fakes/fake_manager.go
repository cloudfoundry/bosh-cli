package fakes

import (
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
)

type CreateInput struct {
	Size            int
	CloudProperties map[string]interface{}
	InstanceID      string
}

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bmdisk.Disk
	CreateErr    error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(
	size int,
	cloudProperties map[string]interface{},
	instanceID string,
) (bmdisk.Disk, error) {
	input := CreateInput{
		Size:            size,
		CloudProperties: cloudProperties,
		InstanceID:      instanceID,
	}
	m.CreateInputs = append(m.CreateInputs, input)

	return m.CreateDisk, m.CreateErr
}
