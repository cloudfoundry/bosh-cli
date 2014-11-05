package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
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

func (m *FakeManager) Create(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (bmvm.CID, error) {
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
