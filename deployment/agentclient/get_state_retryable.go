package agentclient

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/deployment/retrystrategy"
)

type getStateRetryable struct {
	agentClient AgentClient
}

func NewGetStateRetryable(agentClient AgentClient) bmretrystrategy.Retryable {
	return &getStateRetryable{
		agentClient: agentClient,
	}
}

func (r *getStateRetryable) Attempt() (bool, error) {
	stateResponse, err := r.agentClient.GetState()
	if err != nil {
		return false, err
	}

	if stateResponse.JobState == "running" {
		return true, nil
	}

	return true, bosherr.Errorf("Received non-running job state: '%s'", stateResponse.JobState)
}
