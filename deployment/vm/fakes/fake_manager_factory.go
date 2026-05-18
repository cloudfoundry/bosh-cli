package fakes

import (
	"fmt"

	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bitestutils "github.com/cloudfoundry/bosh-cli/v7/testutils"
)

type NewManagerInput struct {
	Cloud      bicloud.Cloud
	DirectorID string
	MbusURL    string
	CACert     string
}

type newManagerOutput struct {
	manager bivm.Manager
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

func (f *FakeManagerFactory) NewManager(cloud bicloud.Cloud, agentClientFactory bihttpagent.AgentClientFactory, directorID, mbusURL, caCert string) bivm.Manager {
	input := NewManagerInput{
		Cloud:      cloud,
		DirectorID: directorID,
		MbusURL:    mbusURL,
		CACert:     caCert,
	}
	f.NewManagerInputs = append(f.NewManagerInputs, input)

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	output, found := f.newManagerBehavior[inputString]
	if !found {
		panic(fmt.Errorf("Unsupported NewManager Input: %#v\nExpected Behavior: %#v", input, f.newManagerBehavior)) //nolint:staticcheck
	}

	return output.manager
}

func (f *FakeManagerFactory) SetNewManagerBehavior(cloud bicloud.Cloud, directorID, mbusURL, caCert string, manager bivm.Manager) {
	input := NewManagerInput{
		Cloud:      cloud,
		DirectorID: directorID,
		MbusURL:    mbusURL,
		CACert:     caCert,
	}

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	f.newManagerBehavior[inputString] = newManagerOutput{manager: manager}
}
