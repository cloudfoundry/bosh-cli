package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type UploadInput struct {
	TarballPath string
}

type uploadOutput struct {
	stemcell bmstemcell.Stemcell
	cid      bmstemcell.CID
	err      error
}

type FakeManager struct {
	UploadInputs   []UploadInput
	uploadBehavior map[UploadInput]uploadOutput
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		UploadInputs:   []UploadInput{},
		uploadBehavior: map[UploadInput]uploadOutput{},
	}
}

func (m *FakeManager) Upload(tarballPath string) (bmstemcell.Stemcell, bmstemcell.CID, error) {
	input := UploadInput{
		TarballPath: tarballPath,
	}
	m.UploadInputs = append(m.UploadInputs, input)
	output, found := m.uploadBehavior[input]
	if !found {
		return bmstemcell.Stemcell{}, bmstemcell.CID(""), fmt.Errorf("Unsupported Upload Input: %s", tarballPath)
	}

	return output.stemcell, output.cid, output.err
}

func (m *FakeManager) SetUploadBehavior(tarballPath string, stemcell bmstemcell.Stemcell, cid bmstemcell.CID, err error) {
	input := UploadInput{
		TarballPath: tarballPath,
	}
	m.uploadBehavior[input] = uploadOutput{stemcell: stemcell, cid: cid, err: err}
}
