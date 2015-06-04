package fakes

import (
	"fmt"

	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	bitempcomp "github.com/cloudfoundry/bosh-init/templatescompiler"
	bitestutils "github.com/cloudfoundry/bosh-init/testutils"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type SaveInput struct {
	Job    bireljob.Job
	Record bitempcomp.TemplateRecord
}

type saveOutput struct {
	err error
}

type FindInput struct {
	Job bireljob.Job
}

type findOutput struct {
	record bitempcomp.TemplateRecord
	found  bool
	err    error
}

type FakeTemplatesRepo struct {
	SaveInputs []SaveInput
	FindInputs []FindInput

	saveBehavior map[string]saveOutput
	findBehavior map[string]findOutput
}

func NewFakeTemplatesRepo() *FakeTemplatesRepo {
	return &FakeTemplatesRepo{
		SaveInputs:   []SaveInput{},
		FindInputs:   []FindInput{},
		saveBehavior: map[string]saveOutput{},
		findBehavior: map[string]findOutput{},
	}
}

func (f *FakeTemplatesRepo) Save(job bireljob.Job, record bitempcomp.TemplateRecord) error {
	input := SaveInput{Job: job, Record: record}
	f.SaveInputs = append(f.SaveInputs, input)

	inputString, err := bitestutils.MarshalToString(input)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling Save input")
	}
	output, found := f.saveBehavior[inputString]

	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Input: Save('%#v', '%#v')", job, record)
}

func (f *FakeTemplatesRepo) SetSaveBehavior(job bireljob.Job, record bitempcomp.TemplateRecord, err error) error {
	input := SaveInput{Job: job, Record: record}
	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}
	f.saveBehavior[inputString] = saveOutput{err: err}
	return nil
}

func (f *FakeTemplatesRepo) Find(job bireljob.Job) (bitempcomp.TemplateRecord, bool, error) {
	input := FindInput{Job: job}
	f.FindInputs = append(f.FindInputs, input)

	inputString, err := bitestutils.MarshalToString(input)
	if err != nil {
		return bitempcomp.TemplateRecord{}, false, bosherr.WrapError(err, "Marshaling Find input")
	}
	output, found := f.findBehavior[inputString]

	if found {
		return output.record, output.found, output.err
	}
	return bitempcomp.TemplateRecord{}, false, fmt.Errorf("Unsupported input: Find('%#v')", job)
}

func (f *FakeTemplatesRepo) SetFindBehavior(job bireljob.Job, record bitempcomp.TemplateRecord, found bool, err error) error {
	input := FindInput{Job: job}
	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}
	f.findBehavior[inputString] = findOutput{record: record, found: found, err: err}
	return nil
}
