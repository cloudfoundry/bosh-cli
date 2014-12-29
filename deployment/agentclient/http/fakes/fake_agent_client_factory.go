package fakes

import (
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
)

type FakeAgentClientFactory struct {
	CreateAgentClient bmagentclient.AgentClient
	CreateDirectorID  string
	CreateMbusURL     string
}

func NewFakeAgentClientFactory() *FakeAgentClientFactory {
	return &FakeAgentClientFactory{}
}

func (f *FakeAgentClientFactory) NewAgentClient(directorID, mbusURL string) bmagentclient.AgentClient {
	f.CreateDirectorID = directorID
	f.CreateMbusURL = mbusURL
	return f.CreateAgentClient
}
