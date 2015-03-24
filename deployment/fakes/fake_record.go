package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type FakeRecord struct {
	IsDeployedInputs   []IsDeployedInput
	isDeployedBehavior map[string]isDeployedOutput

	UpdateInputs   []UpdateInput
	updateBehavior map[string]updateOutput
}

type IsDeployedInput struct {
	ManifestPath string
	Releases     []bmrel.Release
	Stemcell     bmstemcell.ExtractedStemcell
}

type isDeployedOutput struct {
	isDeployed bool
	err        error
}

type UpdateInput struct {
	ManifestPath string
	Releases     []bmrel.Release
}

type updateOutput struct {
	err error
}

func NewFakeRecord() *FakeRecord {
	return &FakeRecord{
		isDeployedBehavior: make(map[string]isDeployedOutput),
		updateBehavior:     make(map[string]updateOutput),
	}
}

func (r *FakeRecord) IsDeployed(manifestPath string, releases []bmrel.Release, stemcell bmstemcell.ExtractedStemcell) (bool, error) {
	input := IsDeployedInput{
		ManifestPath: manifestPath,
		Releases:     releases,
		Stemcell:     stemcell,
	}
	r.IsDeployedInputs = append(r.IsDeployedInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return false, bosherr.WrapError(marshalErr, "Marshaling IsDeployed input")
	}

	output, found := r.isDeployedBehavior[inputString]
	if !found {
		return false, fmt.Errorf("Unsupported IsDeployed Input: %s\nExpected: %#v", inputString, r.isDeployedBehavior)
	}

	return output.isDeployed, output.err
}

func (r *FakeRecord) Update(manifestPath string, releases []bmrel.Release) error {
	input := UpdateInput{
		ManifestPath: manifestPath,
		Releases:     releases,
	}
	r.UpdateInputs = append(r.UpdateInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Update input")
	}

	output, found := r.updateBehavior[inputString]
	if !found {
		return fmt.Errorf("Unsupported Update Input: %s\nExpected: %#v", inputString, r.updateBehavior)
	}

	return output.err
}

func (r *FakeRecord) SetIsDeployedBehavior(
	manifestPath string,
	releases []bmrel.Release,
	stemcell bmstemcell.ExtractedStemcell,
	isDeployed bool,
	err error,
) error {
	input := IsDeployedInput{
		ManifestPath: manifestPath,
		Releases:     releases,
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

func (r *FakeRecord) SetUpdateBehavior(
	manifestPath string,
	releases []bmrel.Release,
	err error,
) error {
	input := UpdateInput{
		ManifestPath: manifestPath,
		Releases:     releases,
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
