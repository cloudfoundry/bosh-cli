package http

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmac "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmhttpclient "github.com/cloudfoundry/bosh-micro-cli/deployment/httpclient"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/deployment/retrystrategy"
)

type agentClient struct {
	agentRequest agentRequest
	getTaskDelay time.Duration
	logger       boshlog.Logger
	logTag       string
}

func NewAgentClient(
	endpoint string,
	directorID string,
	getTaskDelay time.Duration,
	httpClient bmhttpclient.HTTPClient,
	logger boshlog.Logger,
) bmac.AgentClient {
	// if this were NATS, we would need the agentID, but since it's http, the endpoint is unique to the agent
	agentEndpoint := fmt.Sprintf("%s/agent", endpoint)
	agentRequest := NewAgentRequest(agentEndpoint, httpClient, directorID)

	return &agentClient{
		agentRequest: agentRequest,
		getTaskDelay: getTaskDelay,
		logger:       logger,
		logTag:       "httpAgentClient",
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
		return bosherr.Errorf("Failed to start agent services with response: '%s'", response)
	}

	return nil
}

func (c *agentClient) GetState() (bmac.AgentState, error) {
	var response StateResponse
	err := c.agentRequest.Send("get_state", []interface{}{}, &response)
	if err != nil {
		return bmac.AgentState{}, bosherr.WrapError(err, "Sending get_state to the agent")
	}

	agentState := bmac.AgentState{
		JobState: response.Value.JobState,
	}
	return agentState, nil
}

func (c *agentClient) ListDisk() ([]string, error) {
	var response ListResponse
	err := c.agentRequest.Send("list_disk", []interface{}{}, &response)
	if err != nil {
		return []string{}, bosherr.WrapError(err, "Sending 'list_disk' to the agent")
	}

	return response.Value, nil
}

func (c *agentClient) MountDisk(diskCID string) error {
	return c.sendAsyncTaskMessage("mount_disk", []interface{}{diskCID})
}

func (c *agentClient) UnmountDisk(diskCID string) error {
	return c.sendAsyncTaskMessage("unmount_disk", []interface{}{diskCID})
}

func (c *agentClient) MigrateDisk() error {
	return c.sendAsyncTaskMessage("migrate_disk", []interface{}{})
}

func (c *agentClient) sendAsyncTaskMessage(method string, arguments []interface{}) error {
	var response TaskResponse
	err := c.agentRequest.Send(method, arguments, &response)
	if err != nil {
		return bosherr.WrapErrorf(err, "Sending '%s' to the agent", method)
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

		return true, bosherr.Errorf("Task %s is still running", method)
	})

	getTaskRetryStrategy := bmretrystrategy.NewUnlimitedRetryStrategy(c.getTaskDelay, getTaskRetryable, c.logger)
	return getTaskRetryStrategy.Try()
}
