package stemcellfakes

import (
	"fmt"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
)

type NewManagerInput struct {
	Cloud bicloud.Cloud
}

type newManagerOutput struct {
	manager bistemcell.Manager
}

type FakeManagerFactory struct {
	NewManagerInputs   []NewManagerInput
	newManagerBehavior map[bicloud.Cloud]newManagerOutput
}

func NewFakeManagerFactory() *FakeManagerFactory {
	return &FakeManagerFactory{
		NewManagerInputs:   []NewManagerInput{},
		newManagerBehavior: map[bicloud.Cloud]newManagerOutput{},
	}
}

func (f *FakeManagerFactory) NewManager(cloud bicloud.Cloud) bistemcell.Manager {
	input := NewManagerInput{
		Cloud: cloud,
	}
	f.NewManagerInputs = append(f.NewManagerInputs, input)

	output, found := f.newManagerBehavior[cloud]
	if !found {
		panic(fmt.Errorf("Unsupported NewManager Input: %#v\nExpected Behavior: %#v", input, f.newManagerBehavior)) //nolint:staticcheck
	}

	return output.manager
}

func (f *FakeManagerFactory) SetNewManagerBehavior(cloud bicloud.Cloud, manager bistemcell.Manager) {
	f.newManagerBehavior[cloud] = newManagerOutput{manager: manager}
}
