package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeValidator struct {
	ValidateError error
}

func NewFakeValidator() *FakeValidator {
	return &FakeValidator{}
}

func (f *FakeValidator) Validate(release bmrel.Release) error {
	return f.ValidateError
}
