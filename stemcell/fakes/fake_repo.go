package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type SaveInput struct {
	Stemcell bmstemcell.Stemcell
	CID      bmstemcell.CID
}

type SaveOutput struct {
	err error
}

type FindInput struct {
	Stemcell bmstemcell.Stemcell
}

type FindOutput struct {
	cid   bmstemcell.CID
	found bool
	err   error
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

func (fr *FakeRepo) Save(stemcell bmstemcell.Stemcell, cid bmstemcell.CID) error {
	input := SaveInput{
		Stemcell: stemcell,
		CID:      cid,
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

func (fr *FakeRepo) SetSaveBehavior(stemcell bmstemcell.Stemcell, cid bmstemcell.CID, err error) error {
	input := SaveInput{
		Stemcell: stemcell,
		CID:      cid,
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

func (fr *FakeRepo) Find(stemcell bmstemcell.Stemcell) (bmstemcell.CID, bool, error) {
	input := FindInput{
		Stemcell: stemcell,
	}
	fr.FindInputs = append(fr.FindInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return "", false, bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := fr.FindBehavior[inputString]
	if !found {
		return "", false, fmt.Errorf("Unsupported Find Input: %s", inputString)
	}

	return output.cid, output.found, output.err
}

func (fr *FakeRepo) SetFindBehavior(stemcell bmstemcell.Stemcell, cid bmstemcell.CID, found bool, err error) error {
	input := FindInput{
		Stemcell: stemcell,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	fr.FindBehavior[inputString] = FindOutput{
		cid:   cid,
		found: found,
		err:   err,
	}

	return nil
}
