package agentclient

import (
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/deployer/retrystrategy"
)

type pingRetryable struct {
	agentClient AgentClient
}

func NewPingRetryable(agentClient AgentClient) bmretrystrategy.Retryable {
	return &pingRetryable{
		agentClient: agentClient,
	}
}

func (r *pingRetryable) Attempt() (bool, error) {
	_, err := r.agentClient.Ping()
	return true, err
}
