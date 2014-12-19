package http

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmhttpclient "github.com/cloudfoundry/bosh-micro-cli/deployment/httpclient"
)

type AgentClientFactory interface {
	Create(mbusURL string) bmac.AgentClient
}

type agentClientFactory struct {
	deploymentUUID string
	getTaskDelay   time.Duration
	logger         boshlog.Logger
}

func NewAgentClientFactory(
	deploymentUUID string,
	getTaskDelay time.Duration,
	logger boshlog.Logger,
) AgentClientFactory {
	return agentClientFactory{
		deploymentUUID: deploymentUUID,
		getTaskDelay:   getTaskDelay,
		logger:         logger,
	}
}

func (f agentClientFactory) Create(mbusURL string) bmac.AgentClient {
	httpClient := bmhttpclient.NewHTTPClient(f.logger)
	return NewAgentClient(mbusURL, f.deploymentUUID, f.getTaskDelay, httpClient, f.logger)
}
