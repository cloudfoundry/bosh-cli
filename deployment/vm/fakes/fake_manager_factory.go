package fakes

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-init/cloud"
	bmvm "github.com/cloudfoundry/bosh-init/deployment/vm"
	bmtestutils "github.com/cloudfoundry/bosh-init/testutils"
)

type NewManagerInput struct {
	Cloud   bmcloud.Cloud
	MbusURL string
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

func (f *FakeManagerFactory) NewManager(cloud bmcloud.Cloud, mbusURL string) bmvm.Manager {
	input := NewManagerInput{
		Cloud:   cloud,
		MbusURL: mbusURL,
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

func (f *FakeManagerFactory) SetNewManagerBehavior(cloud bmcloud.Cloud, mbusURL string, manager bmvm.Manager) {
	input := NewManagerInput{
		Cloud:   cloud,
		MbusURL: mbusURL,
	}

	inputString, marshalErr := bmtestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	f.newManagerBehavior[inputString] = newManagerOutput{manager: manager}
}
