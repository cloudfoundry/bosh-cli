package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type CreateInput struct {
	StemcellCID bmstemcell.CID
}

type createOutput struct {
	cid bmvm.CID
	err error
}

type FakeInfrastructure struct {
	createBehavior map[string]createOutput
	CreateInputs   []CreateInput
}

func NewFakeInfrastructure() *FakeInfrastructure {
	return &FakeInfrastructure{
		createBehavior: map[string]createOutput{},
		CreateInputs:   []CreateInput{},
	}
}

func (i *FakeInfrastructure) CreateVM(stemcellCID bmstemcell.CID) (bmvm.CID, error) {
	input := CreateInput{StemcellCID: stemcellCID}
	i.CreateInputs = append(i.CreateInputs, input)
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return "", bosherr.WrapError(marshalErr, "Marshaling CreateVM input")
	}

	output, found := i.createBehavior[inputString]
	if found {
		return output.cid, output.err
	}

	return "", fmt.Errorf("Unsupported CreateVM Input: %s\nAvailable inputs: %s", inputString, i.createBehavior)
}

func (i *FakeInfrastructure) SetCreateVMBehavior(stemcellCID bmstemcell.CID, vmCID bmvm.CID, err error) error {
	input := CreateInput{StemcellCID: stemcellCID}
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling CreateVM input")
	}

	i.createBehavior[inputString] = createOutput{cid: vmCID, err: err}
	return nil
}
