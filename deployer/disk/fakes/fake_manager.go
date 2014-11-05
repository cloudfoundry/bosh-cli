package fakes

import (
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
)

type CreateDiskInput struct {
	Size            int
	CloudProperties map[string]interface{}
	InstanceID      string
}

type FakeManager struct {
	CreateDiskInputs []CreateDiskInput
	CreateDiskDisk   bmdisk.Disk
	CreateDiskErr    error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(
	size int,
	cloudProperties map[string]interface{},
	instanceID string,
) (bmdisk.Disk, error) {
	input := CreateDiskInput{
		Size:            size,
		CloudProperties: cloudProperties,
		InstanceID:      instanceID,
	}
	m.CreateDiskInputs = append(m.CreateDiskInputs, input)

	return m.CreateDiskDisk, m.CreateDiskErr
}
