package agentclient

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmhttpclient "github.com/cloudfoundry/bosh-micro-cli/microdeployer/httpclient"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
)

type AgentClient interface {
	Ping() (string, error)
	Stop() error
	Apply(bmas.ApplySpec) error
	Start() error
	GetState() (State, error)
}

type agentClient struct {
	agentRequest agentRequest
	getTaskDelay time.Duration
	logger       boshlog.Logger
	logTag       string
}

func NewAgentClient(
	endpoint string,
	uuid string,
	getTaskDelay time.Duration,
	httpClient bmhttpclient.HTTPClient,
	logger boshlog.Logger,
) AgentClient {
	agentEndpoint := fmt.Sprintf("%s/agent", endpoint)
	agentRequest := NewAgentRequest(agentEndpoint, httpClient, uuid)

	return &agentClient{
		agentRequest: agentRequest,
		getTaskDelay: getTaskDelay,
		logger:       logger,
		logTag:       "agentClient",
	}
}

func (c *agentClient) Ping() (string, error) {
	var response SimpleTaskResponse
	err := c.agentRequest.Send("ping", []interface{}{}, &response)
	if err != nil {
		return "", bosherr.WrapError(err, "Sending ping to the agent")
	}

	return response.Value, nil
}

func (c *agentClient) Stop() error {
	return c.sendAsyncTaskMessage("stop", []interface{}{})
}

func (c *agentClient) Apply(spec bmas.ApplySpec) error {
	return c.sendAsyncTaskMessage("apply", []interface{}{spec})
}

func (c *agentClient) Start() error {
	var response SimpleTaskResponse
	err := c.agentRequest.Send("start", []interface{}{}, &response)
	if err != nil {
		return bosherr.WrapError(err, "Starting agent services")
	}

	if response.Value != "started" {
		return bosherr.New("Failed to start agent services with response: '%s'", response)
	}

	return nil
}

func (c *agentClient) GetState() (State, error) {
	var response StateResponse
	err := c.agentRequest.Send("get_state", []interface{}{}, &response)
	if err != nil {
		return State{}, bosherr.WrapError(err, "Sending get_state to the agent")
	}

	return response.Value, nil
}

func (c *agentClient) sendAsyncTaskMessage(method string, arguments []interface{}) error {
	var response TaskResponse
	err := c.agentRequest.Send(method, arguments, &response)
	if err != nil {
		return bosherr.WrapError(err, "Sending '%s' to the agent", method)
	}

	agentTaskID, err := response.TaskID()
	if err != nil {
		return bosherr.WrapError(err, "Getting agent task id")
	}

	getTaskRetryable := bmretrystrategy.NewRetryable(func() (bool, error) {
		var response TaskResponse
		err = c.agentRequest.Send("get_task", []interface{}{agentTaskID}, &response)
		if err != nil {
			return false, bosherr.WrapError(err, "Sending 'get_task' to the agent")
		}

		taskState, err := response.TaskState()
		if err != nil {
			return false, bosherr.WrapError(err, "Getting task state")
		}

		if taskState != "running" {
			return true, nil
		}

		return true, bosherr.New("Task %s is still running", method)
	})

	getTaskRetryStrategy := bmretrystrategy.NewUnlimitedRetryStrategy(c.getTaskDelay, getTaskRetryable, c.logger)
	return getTaskRetryStrategy.Try()
}
