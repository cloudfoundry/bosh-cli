package agentclient

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmhttpclient "github.com/cloudfoundry/bosh-micro-cli/deployer/httpclient"
)

type Factory interface {
	Create(mbusURL string) AgentClient
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
) Factory {
	return agentClientFactory{
		deploymentUUID: deploymentUUID,
		getTaskDelay:   getTaskDelay,
		logger:         logger,
	}
}

func (f agentClientFactory) Create(mbusURL string) AgentClient {
	httpClient := bmhttpclient.NewHTTPClient(f.logger)
	return NewAgentClient(mbusURL, f.deploymentUUID, f.getTaskDelay, httpClient, f.logger)
}
