package fakes

import (
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type FakeManager struct {
	CreateInputs []CreateInput
	CreateDisk   bidisk.Disk
	CreateErr    error

	findCurrentForVMOutputs map[string]findCurrentForVMOutput

	DeleteUnusedCalledTimes int
	DeleteUnusedErr         error

	findUnusedOutput findUnusedOutput
}

type CreateInput struct {
	DiskPool   bideplmanifest.DiskPool
	InstanceID string
}

type findCurrentForVMOutput struct {
	Disks []bidisk.Disk
	Err   error
}

type findUnusedOutput struct {
	disks []bidisk.Disk
	err   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		findCurrentForVMOutputs: map[string]findCurrentForVMOutput{},
	}
}

func (m *FakeManager) Create(diskPool bideplmanifest.DiskPool, instanceID string) (bidisk.Disk, error) {
	input := CreateInput{
		DiskPool:   diskPool,
		InstanceID: instanceID,
	}
	m.CreateInputs = append(m.CreateInputs, input)

	return m.CreateDisk, m.CreateErr
}

func (m *FakeManager) FindAllCurrent() ([]bidisk.Disk, error) {
	// Default: return all disks set for any VM (aggregated).
	var all []bidisk.Disk
	for _, out := range m.findCurrentForVMOutputs {
		all = append(all, out.Disks...)
	}
	return all, nil
}

func (m *FakeManager) FindCurrentForVM(vmCID string) ([]bidisk.Disk, error) {
	out := m.findCurrentForVMOutputs[vmCID]
	return out.Disks, out.Err
}

func (m *FakeManager) FindUnused() ([]bidisk.Disk, error) {
	return m.findUnusedOutput.disks, m.findUnusedOutput.err
}

func (m *FakeManager) DeleteUnused(eventLogStage biui.Stage) error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
}

func (m *FakeManager) SetFindCurrentForVMBehavior(vmCID string, disks []bidisk.Disk, err error) {
	m.findCurrentForVMOutputs[vmCID] = findCurrentForVMOutput{
		Disks: disks,
		Err:   err,
	}
}

func (m *FakeManager) SetFindUnusedBehavior(disks []bidisk.Disk, err error) {
	m.findUnusedOutput = findUnusedOutput{disks: disks, err: err}
}
