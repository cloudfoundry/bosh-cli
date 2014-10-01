package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type NewManagerInput struct {
	Infrastructure bmvm.Infrastructure
}

type newManagerOutput struct {
	manager bmvm.Manager
}

type FakeManagerFactory struct {
	NewManagerInputs   []NewManagerInput
	newManagerBehavior map[string]newManagerOutput
}

func NewFakeManagerFactory() *FakeManagerFactory {
	return &FakeManagerFactory{
		NewManagerInputs:   []NewManagerInput{},
		newManagerBehavior: map[string]newManagerOutput{},
	}
}

func (f *FakeManagerFactory) NewManager(infrastructure bmvm.Infrastructure) bmvm.Manager {
	input := NewManagerInput{
		Infrastructure: infrastructure,
	}
	f.NewManagerInputs = append(f.NewManagerInputs, input)

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	output, found := f.newManagerBehavior[inputString]
	if !found {
		panic(fmt.Errorf("Unsupported NewManager Input: %#v\nExpected Behavior: %#v", input, f.newManagerBehavior))
	}

	return output.manager
}

func (f *FakeManagerFactory) SetNewManagerBehavior(infrastructure bmvm.Infrastructure, manager bmvm.Manager) {
	input := NewManagerInput{
		Infrastructure: infrastructure,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	f.newManagerBehavior[inputString] = newManagerOutput{manager: manager}
}
