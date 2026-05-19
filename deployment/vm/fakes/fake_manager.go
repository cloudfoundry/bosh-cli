package fakes

import (
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
)

type CreateInput struct {
	JobName    string
	InstanceID int
	AZ         string
	Stemcell   bistemcell.CloudStemcell
	Manifest   bideplmanifest.Manifest
	DiskCIDs   []string
}

type FakeManager struct {
	CreateInput CreateInput
	CreateVM    bivm.VM
	CreateErr   error

	findAllBehaviour findAllOutput
}

type findAllOutput struct {
	vms []bivm.ExistingVM
	err error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) FindAll() ([]bivm.ExistingVM, error) {
	return m.findAllBehaviour.vms, m.findAllBehaviour.err
}

func (m *FakeManager) Create(jobName string, instanceID int, az string, stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest, diskCIDs []string) (bivm.VM, error) {
	m.CreateInput = CreateInput{
		JobName:    jobName,
		InstanceID: instanceID,
		AZ:         az,
		Stemcell:   stemcell,
		Manifest:   deploymentManifest,
		DiskCIDs:   diskCIDs,
	}

	return m.CreateVM, m.CreateErr
}

func (m *FakeManager) SetFindAllBehavior(vms []bivm.ExistingVM, err error) {
	m.findAllBehaviour = findAllOutput{
		vms: vms,
		err: err,
	}
}
