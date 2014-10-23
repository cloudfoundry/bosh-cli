package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type CreateInput struct {
	StemcellManifest bmstemcell.Manifest
}

type createOutput struct {
	cid bmstemcell.CID
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

func (i *FakeInfrastructure) CreateStemcell(stemcellManifest bmstemcell.Manifest) (bmstemcell.CID, error) {
	input := CreateInput{StemcellManifest: stemcellManifest}
	i.CreateInputs = append(i.CreateInputs, input)
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return "", bosherr.WrapError(marshalErr, "Marshaling CreateStemcell input")
	}

	output, found := i.createBehavior[inputString]
	if found {
		return output.cid, output.err
	}

	return "", fmt.Errorf("Unsupported CreateStemcell Input: %s\nAvailable inputs: %s", inputString, i.createBehavior)
}

func (i *FakeInfrastructure) SetCreateStemcellBehavior(stemcellManifest bmstemcell.Manifest, cid bmstemcell.CID, err error) error {
	input := CreateInput{StemcellManifest: stemcellManifest}
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling CreateStemcell input")
	}

	i.createBehavior[inputString] = createOutput{cid: cid, err: err}
	return nil
}
