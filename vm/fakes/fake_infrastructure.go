package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type CreateInput struct {
	StemcellCID  bmstemcell.CID
	NetworksSpec map[string]interface{}
}

type createOutput struct {
	cid bmvm.CID
	err error
}

type FakeInfrastructure struct {
	createOutput createOutput
	CreateInput  CreateInput
}

func NewFakeInfrastructure() *FakeInfrastructure {
	return &FakeInfrastructure{
		createOutput: createOutput{},
		CreateInput:  CreateInput{},
	}
}

func (i *FakeInfrastructure) CreateVM(stemcellCID bmstemcell.CID, networksSpec map[string]interface{}) (bmvm.CID, error) {
	input := CreateInput{
		StemcellCID:  stemcellCID,
		NetworksSpec: networksSpec,
	}
	i.CreateInput = input

	if (i.createOutput != createOutput{}) {
		return i.createOutput.cid, i.createOutput.err
	}

	return "", nil
}

func (i *FakeInfrastructure) SetCreateVMBehavior(vmCID bmvm.CID, err error) error {
	i.createOutput = createOutput{cid: vmCID, err: err}
	return nil
}
