package http

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmhttpclient "github.com/cloudfoundry/bosh-micro-cli/deployment/httpclient"
)

type AgentClientFactory interface {
	NewAgentClient(directorID, mbusURL string) bmagentclient.AgentClient
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

func (f *agentClientFactory) NewAgentClient(directorID, mbusURL string) bmagentclient.AgentClient {
	httpClient := bmhttpclient.NewHTTPClient(f.logger)
	return NewAgentClient(mbusURL, directorID, f.getTaskDelay, httpClient, f.logger)
}
