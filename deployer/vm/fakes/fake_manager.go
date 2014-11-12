package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type CreateInput struct {
	Stemcell   bmstemcell.CloudStemcell
	Deployment bmdepl.Deployment
	MbusURL    string
}

type FakeManager struct {
	CreateInput CreateInput
	CreateVM    bmvm.VM
	CreateErr   error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (m *FakeManager) Create(stemcell bmstemcell.CloudStemcell, deployment bmdepl.Deployment, mbusURL string) (bmvm.VM, error) {
	input := CreateInput{
		Stemcell:   stemcell,
		Deployment: deployment,
		MbusURL:    mbusURL,
	}
	m.CreateInput = input

	return m.CreateVM, m.CreateErr
}
