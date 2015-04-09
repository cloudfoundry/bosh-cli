package fakes

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
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
	Manifest biinstallmanifest.Manifest
}

type ValidateOutput struct {
	Err error
}

func (v *FakeValidator) Validate(manifest biinstallmanifest.Manifest) error {
	v.ValidateInputs = append(v.ValidateInputs, ValidateInput{
		Manifest: manifest,
	})

	if len(v.validateOutputs) == 0 {
		return bosherr.Errorf("Unexpected FakeValidator.Validate(manifest) called with manifest: %#v", manifest)
	}
	validateOutput := v.validateOutputs[0]
	v.validateOutputs = v.validateOutputs[1:]
	return validateOutput.Err
}

func (v *FakeValidator) SetValidateBehavior(outputs []ValidateOutput) {
	v.validateOutputs = outputs
}
