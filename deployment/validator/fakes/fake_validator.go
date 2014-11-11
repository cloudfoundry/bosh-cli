package fakes

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type FakeValidator struct {
	ValidateInputs  []ValidateInput
	validateOutputs []ValidateOutput
}

func NewFakeValidator() *FakeValidator {
	return &FakeValidator{
		ValidateInputs:  []ValidateInput{},
		validateOutputs: []ValidateOutput{},
	}
}

type ValidateInput struct {
	Deployment bmdepl.Deployment
}

type ValidateOutput struct {
	Err error
}

func (v *FakeValidator) Validate(deployment bmdepl.Deployment) error {
	v.ValidateInputs = append(v.ValidateInputs, ValidateInput{
		Deployment: deployment,
	})

	validateOutput := v.validateOutputs[0]
	v.validateOutputs = v.validateOutputs[1:]
	return validateOutput.Err
}

func (v *FakeValidator) SetValidateBehavior(outputs []ValidateOutput) {
	v.validateOutputs = outputs
}
