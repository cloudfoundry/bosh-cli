package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type CreateInput struct {
	Stemcell bmstemcell.CloudStemcell
	Manifest bmdepl.Manifest
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

func (m *FakeManager) Create(stemcell bmstemcell.CloudStemcell, deploymentManifest bmdepl.Manifest) (bmvm.VM, error) {
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
