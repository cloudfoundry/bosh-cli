package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
)

type SaveInput struct {
	Config bmconfig.Config
}

type SaveOutput struct {
	err error
}

type FakeService struct {
	SaveInputs   []SaveInput
	saveBehavior map[string]SaveOutput
}

func NewFakeService() *FakeService {
	return &FakeService{
		SaveInputs:   []SaveInput{},
		saveBehavior: map[string]SaveOutput{},
	}
}

func (s *FakeService) Load() (bmconfig.Config, error) {
	return bmconfig.Config{}, nil
}

func (s *FakeService) Save(config bmconfig.Config) error {
	input := SaveInput{Config: config}
	s.SaveInputs = append(s.SaveInputs, input)

	inputString, err := bmtestutils.MarshalToString(input)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling Save input")
	}
	output, found := s.saveBehavior[inputString]
	if found {
		return output.err
	}
	return fmt.Errorf("Unsupported Save Input: %s", inputString)
}

func (s *FakeService) SetSaveBehavior(config bmconfig.Config, err error) error {
	input := SaveInput{Config: config}
	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}
	s.saveBehavior[inputString] = SaveOutput{err: err}
	return nil
}
