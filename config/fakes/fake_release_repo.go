package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type ReleaseRepoSaveInput struct {
	Name    string
	Version string
}

type ReleaseRepoSaveOutput struct {
	releaseRecord bmconfig.ReleaseRecord
	err           error
}

type ReleaseRepoFindInput struct {
	Name    string
	Version string
}

type ReleaseRepoFindOutput struct {
	releaseRecord bmconfig.ReleaseRecord
	found         bool
	err           error
}

type ReleaseRepoFindCurrentOutput struct {
	releaseRecord bmconfig.ReleaseRecord
	found         bool
	err           error
}

type FakeReleaseRepo struct {
	SaveBehavior map[string]ReleaseRepoSaveOutput
	SaveInputs   []ReleaseRepoSaveInput
	FindBehavior map[string]ReleaseRepoFindOutput
	FindInputs   []ReleaseRepoFindInput

	UpdateCurrentRecordID string
	UpdateCurrentErr      error

	findCurrentOutput ReleaseRepoFindCurrentOutput
}

func NewFakeReleaseRepo() *FakeReleaseRepo {
	return &FakeReleaseRepo{
		FindBehavior: map[string]ReleaseRepoFindOutput{},
		FindInputs:   []ReleaseRepoFindInput{},
		SaveBehavior: map[string]ReleaseRepoSaveOutput{},
		SaveInputs:   []ReleaseRepoSaveInput{},
	}
}

func (fr *FakeReleaseRepo) UpdateCurrent(recordID string) error {
	fr.UpdateCurrentRecordID = recordID
	return fr.UpdateCurrentErr
}

func (fr *FakeReleaseRepo) FindCurrent() (bmconfig.ReleaseRecord, bool, error) {
	return fr.findCurrentOutput.releaseRecord, fr.findCurrentOutput.found, fr.findCurrentOutput.err
}

func (fr *FakeReleaseRepo) Save(name, version string) (bmconfig.ReleaseRecord, error) {
	input := ReleaseRepoSaveInput{
		Name:    name,
		Version: version,
	}
	fr.SaveInputs = append(fr.SaveInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bmconfig.ReleaseRecord{}, bosherr.WrapError(marshalErr, "Marshaling Save input")
	}

	output, found := fr.SaveBehavior[inputString]
	if !found {
		return bmconfig.ReleaseRecord{}, fmt.Errorf("Unsupported Save Input: %s", inputString)
	}

	return output.releaseRecord, output.err
}

func (fr *FakeReleaseRepo) SetSaveBehavior(name, version string, releaseRecord bmconfig.ReleaseRecord, err error) error {
	input := ReleaseRepoSaveInput{
		Name:    name,
		Version: version,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}

	fr.SaveBehavior[inputString] = ReleaseRepoSaveOutput{
		releaseRecord: releaseRecord,
		err:           err,
	}

	return nil
}

func (fr *FakeReleaseRepo) Find(name, version string) (bmconfig.ReleaseRecord, bool, error) {
	input := ReleaseRepoFindInput{
		Name:    name,
		Version: version,
	}
	fr.FindInputs = append(fr.FindInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bmconfig.ReleaseRecord{}, false, bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := fr.FindBehavior[inputString]
	if !found {
		return bmconfig.ReleaseRecord{}, false, fmt.Errorf("Unsupported Find Input: %s", inputString)
	}

	return output.releaseRecord, output.found, output.err
}

func (fr *FakeReleaseRepo) SetFindBehavior(name, version string, foundRecord bmconfig.ReleaseRecord, found bool, err error) error {
	input := ReleaseRepoFindInput{
		Name:    name,
		Version: version,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	fr.FindBehavior[inputString] = ReleaseRepoFindOutput{
		releaseRecord: foundRecord,
		found:         found,
		err:           err,
	}

	return nil
}

func (fr *FakeReleaseRepo) SetFindCurrentBehavior(foundRecord bmconfig.ReleaseRecord, found bool, err error) error {
	fr.findCurrentOutput = ReleaseRepoFindCurrentOutput{
		releaseRecord: foundRecord,
		found:         found,
		err:           err,
	}

	return nil
}
