package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type CreateInput struct {
	StemcellCID bmstemcell.CID
	Deployment  bmdepl.Deployment
}

type FakeManager struct {
	CreateInput CreateInput
	CreateVM    bmvm.VM
	CreateErr   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (bmvm.VM, error) {
	input := CreateInput{
		StemcellCID: stemcellCID,
		Deployment:  deployment,
	}
	m.CreateInput = input

	return m.CreateVM, m.CreateErr
}
