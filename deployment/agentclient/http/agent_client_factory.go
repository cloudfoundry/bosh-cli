package http

import (
	"time"

	biagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
	bihttpclient "github.com/cloudfoundry/bosh-init/deployment/httpclient"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
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
	httpClient := bihttpclient.NewHTTPClient(bihttpclient.DefaultClient, f.logger)
	return NewAgentClient(mbusURL, directorID, f.getTaskDelay, httpClient, f.logger)
}
