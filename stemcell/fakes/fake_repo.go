package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type SaveInput struct {
	stemcellPath string
}

type SaveOutput struct {
	stemcell bmstemcell.Stemcell
	destPath string
	err      error
}

type FakeRepo struct {
	SaveBehavior map[SaveInput]SaveOutput
}

func NewFakeRepo() *FakeRepo {
	return &FakeRepo{SaveBehavior: map[SaveInput]SaveOutput{}}
}

func (fr *FakeRepo) Save(stemcellPath string) (bmstemcell.Stemcell, string, error) {
	input := SaveInput{
		stemcellPath: stemcellPath,
	}
	output, found := fr.SaveBehavior[input]
	if !found {
		return bmstemcell.Stemcell{}, "", fmt.Errorf("Unsupported Input: Save('%#v')", stemcellPath)
	}

	return output.stemcell, output.destPath, output.err
}

func (fr *FakeRepo) SetSaveBehavior(stemcellPath, destPath string, stemcell bmstemcell.Stemcell, err error) {
	input := SaveInput{
		stemcellPath: stemcellPath,
	}
	fr.SaveBehavior[input] = SaveOutput{
		stemcell: stemcell,
		destPath: destPath,
		err:      err,
	}
}
