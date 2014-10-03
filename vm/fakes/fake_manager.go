package fakes

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type CreateVMInput struct {
	StemcellCID bmstemcell.CID
	Deployment  bmdepl.Deployment
}

type createVMOutput struct {
	cid bmvm.CID
	err error
}

type FakeManager struct {
	CreateVMInput  CreateVMInput
	CreateVMOutput createVMOutput
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		CreateVMInput: CreateVMInput{},
	}
}

func (m *FakeManager) CreateVM(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (bmvm.CID, error) {
	input := CreateVMInput{
		StemcellCID: stemcellCID,
		Deployment:  deployment,
	}
	m.CreateVMInput = input

	if (m.CreateVMOutput != createVMOutput{}) {
		return m.CreateVMOutput.cid, m.CreateVMOutput.err
	}

	return "", nil
}

func (m *FakeManager) SetCreateVMBehavior(cid bmvm.CID, err error) {
	m.CreateVMOutput = createVMOutput{cid: cid, err: err}
}
