package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
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
	Stage    bmui.Stage
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

func (m *FakeManager) FindCurrent() ([]bmstemcell.CloudStemcell, error) {
	return []bmstemcell.CloudStemcell{}, bosherr.Error("FakeManager.FindCurrent() not implemented (yet)")
}

func (m *FakeManager) Upload(stemcell bmstemcell.ExtractedStemcell, stage bmui.Stage) (bmstemcell.CloudStemcell, error) {
	input := UploadInput{
		Stemcell: stemcell,
		Stage:    stage,
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

func (m *FakeManager) DeleteUnused(eventLoggerStage bmui.Stage) error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
}

func (m *FakeManager) SetUploadBehavior(
	extractedStemcell bmstemcell.ExtractedStemcell,
	stage bmui.Stage,
	cloudStemcell bmstemcell.CloudStemcell,
	err error,
) {
	input := UploadInput{
		Stemcell: extractedStemcell,
		Stage:    stage,
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
