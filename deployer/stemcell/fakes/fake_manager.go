package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
)

type ExtractInput struct {
	TarballPath string
}

type UploadInput struct {
	Stemcell bmstemcell.ExtractedStemcell
}

type extractOutput struct {
	stemcell bmstemcell.ExtractedStemcell
	err      error
}

type uploadOutput struct {
	stemcell bmstemcell.CloudStemcell
	err      error
}

type FakeManager struct {
	ExtractInputs   []ExtractInput
	extractBehavior map[ExtractInput]extractOutput

	UploadInputs   []UploadInput
	uploadBehavior map[UploadInput]uploadOutput
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		UploadInputs:   []UploadInput{},
		uploadBehavior: map[UploadInput]uploadOutput{},

		ExtractInputs:   []ExtractInput{},
		extractBehavior: map[ExtractInput]extractOutput{},
	}
}

func (m *FakeManager) Extract(tarballPath string) (bmstemcell.ExtractedStemcell, error) {
	input := ExtractInput{
		TarballPath: tarballPath,
	}
	m.ExtractInputs = append(m.ExtractInputs, input)
	output, found := m.extractBehavior[input]
	if !found {
		return nil, fmt.Errorf("Unsupported Upload Input: %s", tarballPath)
	}

	return output.stemcell, output.err
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

func (m *FakeManager) SetExtractBehavior(
	tarballPath string,
	extractedStemcell bmstemcell.ExtractedStemcell,
	err error,
) {
	input := ExtractInput{
		TarballPath: tarballPath,
	}
	m.extractBehavior[input] = extractOutput{stemcell: extractedStemcell, err: err}
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
