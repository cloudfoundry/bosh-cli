package fakes

import (
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
)

type CreateInput struct {
	Stemcell bmstemcell.CloudStemcell
	Manifest bmmanifest.Manifest
}

type FakeManager struct {
	CreateInput CreateInput
	CreateVM    bmvm.VM
	CreateErr   error

	findCurrentBehaviour findCurrentOutput
}

type findCurrentOutput struct {
	vm    bmvm.VM
	found bool
	err   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) FindCurrent() (bmvm.VM, bool, error) {
	return m.findCurrentBehaviour.vm, m.findCurrentBehaviour.found, m.findCurrentBehaviour.err
}

func (m *FakeManager) Create(stemcell bmstemcell.CloudStemcell, deploymentManifest bmmanifest.Manifest) (bmvm.VM, error) {
	input := CreateInput{
		Stemcell: stemcell,
		Manifest: deploymentManifest,
	}
	m.CreateInput = input

	return m.CreateVM, m.CreateErr
}

func (m *FakeManager) SetFindCurrentBehavior(vm bmvm.VM, found bool, err error) {
	m.findCurrentBehaviour = findCurrentOutput{
		vm:    vm,
		found: found,
		err:   err,
	}
}
