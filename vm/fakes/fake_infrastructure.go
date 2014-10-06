package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type CreateInput struct {
	StemcellCID     bmstemcell.CID
	NetworksSpec    map[string]interface{}
	CloudProperties map[string]interface{}
	Env             map[string]interface{}
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

func (i *FakeInfrastructure) CreateVM(
	stemcellCID bmstemcell.CID,
	cloudProperties map[string]interface{},
	networksSpec map[string]interface{},
	env map[string]interface{},
) (bmvm.CID, error) {
	input := CreateInput{
		StemcellCID:     stemcellCID,
		CloudProperties: cloudProperties,
		NetworksSpec:    networksSpec,
		Env:             env,
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
