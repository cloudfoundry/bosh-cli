package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type SaveInput struct {
	StemcellManifest bmstemcell.Manifest
	Stemcell         bmstemcell.CloudStemcell
}

type SaveOutput struct {
	err error
}

type FindInput struct {
	StemcellManifest bmstemcell.Manifest
}

type FindOutput struct {
	stemcell bmstemcell.CloudStemcell
	found    bool
	err      error
}

type FakeRepo struct {
	SaveBehavior map[string]SaveOutput
	SaveInputs   []SaveInput
	FindBehavior map[string]FindOutput
	FindInputs   []FindInput
}

func NewFakeRepo() *FakeRepo {
	return &FakeRepo{
		FindBehavior: map[string]FindOutput{},
		FindInputs:   []FindInput{},
		SaveBehavior: map[string]SaveOutput{},
		SaveInputs:   []SaveInput{},
	}
}

func (fr *FakeRepo) Save(stemcellManifest bmstemcell.Manifest, stemcell bmstemcell.CloudStemcell) error {
	input := SaveInput{
		StemcellManifest: stemcellManifest,
		Stemcell:         stemcell,
	}
	fr.SaveInputs = append(fr.SaveInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}

	output, found := fr.SaveBehavior[inputString]
	if !found {
		return fmt.Errorf("Unsupported Save Input: %s", inputString)
	}

	return output.err
}

func (fr *FakeRepo) SetSaveBehavior(stemcellManifest bmstemcell.Manifest, stemcell bmstemcell.CloudStemcell, err error) error {
	input := SaveInput{
		StemcellManifest: stemcellManifest,
		Stemcell:         stemcell,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}

	fr.SaveBehavior[inputString] = SaveOutput{
		err: err,
	}

	return nil
}

func (fr *FakeRepo) Find(stemcellManifest bmstemcell.Manifest) (bmstemcell.CloudStemcell, bool, error) {
	input := FindInput{
		StemcellManifest: stemcellManifest,
	}
	fr.FindInputs = append(fr.FindInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bmstemcell.CloudStemcell{}, false, bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := fr.FindBehavior[inputString]
	if !found {
		return bmstemcell.CloudStemcell{}, false, fmt.Errorf("Unsupported Find Input: %s", inputString)
	}

	return output.stemcell, output.found, output.err
}

func (fr *FakeRepo) SetFindBehavior(stemcellManifest bmstemcell.Manifest, stemcell bmstemcell.CloudStemcell, found bool, err error) error {
	input := FindInput{
		StemcellManifest: stemcellManifest,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	fr.FindBehavior[inputString] = FindOutput{
		stemcell: stemcell,
		found:    found,
		err:      err,
	}

	return nil
}
