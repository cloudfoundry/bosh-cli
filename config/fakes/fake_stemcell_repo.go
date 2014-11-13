package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type SaveInput struct {
	StemcellRecord bmconfig.StemcellRecord
}

type SaveOutput struct {
	err error
}

type FindInput struct {
	Name    string
	Version string
}

type FindOutput struct {
	stemcellRecord bmconfig.StemcellRecord
	found          bool
	err            error
}

type FakeStemcellRepo struct {
	SaveBehavior map[string]SaveOutput
	SaveInputs   []SaveInput
	FindBehavior map[string]FindOutput
	FindInputs   []FindInput
}

func NewFakeStemcellRepo() *FakeStemcellRepo {
	return &FakeStemcellRepo{
		FindBehavior: map[string]FindOutput{},
		FindInputs:   []FindInput{},
		SaveBehavior: map[string]SaveOutput{},
		SaveInputs:   []SaveInput{},
	}
}

func (fr *FakeStemcellRepo) Save(stemcellRecord bmconfig.StemcellRecord) error {
	input := SaveInput{
		StemcellRecord: stemcellRecord,
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

func (fr *FakeStemcellRepo) SetSaveBehavior(stemcellRecord bmconfig.StemcellRecord, err error) error {
	input := SaveInput{
		StemcellRecord: stemcellRecord,
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

func (fr *FakeStemcellRepo) Find(name, version string) (bmconfig.StemcellRecord, bool, error) {
	input := FindInput{
		Name:    name,
		Version: version,
	}
	fr.FindInputs = append(fr.FindInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bmconfig.StemcellRecord{}, false, bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := fr.FindBehavior[inputString]
	if !found {
		return bmconfig.StemcellRecord{}, false, fmt.Errorf("Unsupported Find Input: %s", inputString)
	}

	return output.stemcellRecord, output.found, output.err
}

func (fr *FakeStemcellRepo) SetFindBehavior(name, version string, foundRecord bmconfig.StemcellRecord, found bool, err error) error {
	input := FindInput{
		Name:    name,
		Version: version,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	fr.FindBehavior[inputString] = FindOutput{
		stemcellRecord: foundRecord,
		found:          found,
		err:            err,
	}

	return nil
}
