package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type FakeDeploymentRecord struct {
	IsDeployedInputs   []IsDeployedInput
	isDeployedBehavior map[string]isDeployedOutput

	UpdateInputs   []UpdateInput
	updateBehavior map[string]updateOutput
}

type IsDeployedInput struct {
	ManifestPath string
	Release      bmrel.Release
	Stemcell     bmstemcell.ExtractedStemcell
}

type isDeployedOutput struct {
	isDeployed bool
	err        error
}

type UpdateInput struct {
	ManifestPath string
	Release      bmrel.Release
	Stemcell     bmstemcell.ExtractedStemcell
}

type updateOutput struct {
	err error
}

func NewFakeDeploymentRecord() *FakeDeploymentRecord {
	return &FakeDeploymentRecord{
		isDeployedBehavior: make(map[string]isDeployedOutput),
		updateBehavior:     make(map[string]updateOutput),
	}
}

func (r *FakeDeploymentRecord) IsDeployed(manifestPath string, release bmrel.Release, stemcell bmstemcell.ExtractedStemcell) (bool, error) {
	input := IsDeployedInput{
		ManifestPath: manifestPath,
		Release:      release,
		Stemcell:     stemcell,
	}
	r.IsDeployedInputs = append(r.IsDeployedInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return false, bosherr.WrapError(marshalErr, "Marshaling IsDeployed input")
	}

	output, found := r.isDeployedBehavior[inputString]
	if !found {
		return false, fmt.Errorf("Unsupported Find Input: %s\nExpected: %#v", inputString, r.isDeployedBehavior)
	}

	return output.isDeployed, output.err
}

func (r *FakeDeploymentRecord) Update(manifestPath string, release bmrel.Release, stemcell bmstemcell.ExtractedStemcell) error {
	input := UpdateInput{
		ManifestPath: manifestPath,
		Release:      release,
		Stemcell:     stemcell,
	}
	r.UpdateInputs = append(r.UpdateInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Update input")
	}

	output, found := r.isDeployedBehavior[inputString]
	if !found {
		return fmt.Errorf("Unsupported Find Input: %s", inputString)
	}

	return output.err
}

func (r *FakeDeploymentRecord) SetIsDeployedBehavior(
	manifestPath string,
	release bmrel.Release,
	stemcell bmstemcell.ExtractedStemcell,
	isDeployed bool,
	err error,
) error {
	input := IsDeployedInput{
		ManifestPath: manifestPath,
		Release:      release,
		Stemcell:     stemcell,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling IsDeployed input")
	}

	r.isDeployedBehavior[inputString] = isDeployedOutput{
		isDeployed: isDeployed,
		err:        err,
	}

	return nil
}

func (r *FakeDeploymentRecord) SetUpdateBehavior(
	manifestPath string,
	release bmrel.Release,
	stemcell bmstemcell.ExtractedStemcell,
	err error,
) error {
	input := UpdateInput{
		ManifestPath: manifestPath,
		Release:      release,
		Stemcell:     stemcell,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Update input")
	}

	r.updateBehavior[inputString] = updateOutput{
		err: err,
	}

	return nil
}
