package http

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	biagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
	bihttpclient "github.com/cloudfoundry/bosh-init/deployment/httpclient"
)

type AgentClientFactory interface {
	NewAgentClient(directorID, mbusURL string) biagentclient.AgentClient
}

type agentClientFactory struct {
	getTaskDelay time.Duration
	logger       boshlog.Logger
}

func NewAgentClientFactory(
	getTaskDelay time.Duration,
	logger boshlog.Logger,
) AgentClientFactory {
	return &agentClientFactory{
		getTaskDelay: getTaskDelay,
		logger:       logger,
	}
}

func (f *agentClientFactory) NewAgentClient(directorID, mbusURL string) biagentclient.AgentClient {
	httpClient := bihttpclient.NewHTTPClient(f.logger)
	return NewAgentClient(mbusURL, directorID, f.getTaskDelay, httpClient, f.logger)
}
