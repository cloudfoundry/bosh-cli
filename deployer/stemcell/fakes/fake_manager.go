package fakes

import (
	"fmt"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeManager struct {
	UploadInputs   []UploadInput
	uploadBehavior map[UploadInput]uploadOutput

	findUnusedOutput findUnusedOutput

	DeleteUnusedCalledTimes int
	DeleteUnusedErr         error
}

type UploadInput struct {
	Stemcell bmstemcell.ExtractedStemcell
}

type uploadOutput struct {
	stemcell bmstemcell.CloudStemcell
	err      error
}

type findUnusedOutput struct {
	stemcells []bmstemcell.CloudStemcell
	err       error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		UploadInputs:   []UploadInput{},
		uploadBehavior: map[UploadInput]uploadOutput{},
	}
}

func (m *FakeManager) FindCurrent() (bmstemcell.CloudStemcell, bool, error) {
	return nil, false, nil
}

func (m *FakeManager) Upload(stemcell bmstemcell.ExtractedStemcell) (bmstemcell.CloudStemcell, error) {
	input := UploadInput{
		Stemcell: stemcell,
	}
	m.UploadInputs = append(m.UploadInputs, input)
	output, found := m.uploadBehavior[input]
	if !found {
		return nil, fmt.Errorf("Unsupported Upload Input: %#v", stemcell)
	}

	return output.stemcell, output.err
}

func (m *FakeManager) FindUnused() ([]bmstemcell.CloudStemcell, error) {
	return m.findUnusedOutput.stemcells, m.findUnusedOutput.err
}

func (m *FakeManager) DeleteUnused(eventLoggerStage bmeventlog.Stage) error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
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

func (m *FakeManager) SetFindUnusedBehavior(
	stemcells []bmstemcell.CloudStemcell,
	err error,
) {
	m.findUnusedOutput = findUnusedOutput{
		stemcells: stemcells,
		err:       err,
	}
}
