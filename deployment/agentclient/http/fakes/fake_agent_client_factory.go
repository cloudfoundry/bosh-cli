package fakes

import (
	biagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
)

type FakeAgentClientFactory struct {
	CreateAgentClient biagentclient.AgentClient
	CreateDirectorID  string
	CreateMbusURL     string
}

func NewFakeAgentClientFactory() *FakeAgentClientFactory {
	return &FakeAgentClientFactory{}
}

func (f *FakeAgentClientFactory) NewAgentClient(directorID, mbusURL string) biagentclient.AgentClient {
	f.CreateDirectorID = directorID
	f.CreateMbusURL = mbusURL
	return f.CreateAgentClient
}
