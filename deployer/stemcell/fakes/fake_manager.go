package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
)

type UploadInput struct {
	Stemcell bmstemcell.ExtractedStemcell
}

type uploadOutput struct {
	stemcell bmstemcell.CloudStemcell
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

func (m *FakeManager) Upload(stemcell bmstemcell.ExtractedStemcell) (bmstemcell.CloudStemcell, error) {
	input := UploadInput{
		Stemcell: stemcell,
	}
	m.UploadInputs = append(m.UploadInputs, input)
	output, found := m.uploadBehavior[input]
	if !found {
		return bmstemcell.CloudStemcell{}, fmt.Errorf("Unsupported Upload Input: %#v", stemcell)
	}

	return output.stemcell, output.err
}

func (m *FakeManager) SetUploadBehavior(
	extractedStemcell bmstemcell.ExtractedStemcell,
	cloudStemcell bmstemcell.CloudStemcell,
	err error,
) {
	input := UploadInput{
		Stemcell: extractedStemcell,
	}
	m.uploadBehavior[input] = uploadOutput{stemcell: cloudStemcell, err: err}
}
