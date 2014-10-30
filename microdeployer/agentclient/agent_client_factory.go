package agentclient

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type Factory interface {
	Create(string) AgentClient
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
	return NewAgentClient(mbusURL, f.deploymentUUID, f.getTaskDelay, f.logger)
}
