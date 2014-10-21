package agentclient

import (
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
)

type pingRetryable struct {
	agentClient AgentClient
}

func NewPingRetryable(agentClient AgentClient) bmretrystrategy.Retryable {
	return &pingRetryable{
		agentClient: agentClient,
	}
}

func (r *pingRetryable) Attempt() error {
	_, err := r.agentClient.Ping()
	return err
}
