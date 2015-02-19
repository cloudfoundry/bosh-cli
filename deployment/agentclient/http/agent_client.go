package http

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshretry "github.com/cloudfoundry/bosh-agent/retrystrategy"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmhttpclient "github.com/cloudfoundry/bosh-micro-cli/deployment/httpclient"
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
) bmagentclient.AgentClient {
	// if this were NATS, we would need the agentID, but since it's http, the endpoint is unique to the agent
	agentEndpoint := fmt.Sprintf("%s/agent", endpoint)
	agentRequest := agentRequest{
		directorID: directorID,
		endpoint:   agentEndpoint,
		httpClient: httpClient,
	}
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
	_, err := c.sendAsyncTaskMessage("stop", []interface{}{})
	return err
}

func (c *agentClient) Apply(spec bmas.ApplySpec) error {
	_, err := c.sendAsyncTaskMessage("apply", []interface{}{spec})
	return err
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

func (c *agentClient) GetState() (bmagentclient.AgentState, error) {
	var response StateResponse
	err := c.agentRequest.Send("get_state", []interface{}{}, &response)
	if err != nil {
		return bmagentclient.AgentState{}, bosherr.WrapError(err, "Sending get_state to the agent")
	}

	agentState := bmagentclient.AgentState{
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
	_, err := c.sendAsyncTaskMessage("mount_disk", []interface{}{diskCID})
	return err
}

func (c *agentClient) UnmountDisk(diskCID string) error {
	_, err := c.sendAsyncTaskMessage("unmount_disk", []interface{}{diskCID})
	return err
}

func (c *agentClient) MigrateDisk() error {
	_, err := c.sendAsyncTaskMessage("migrate_disk", []interface{}{})
	return err
}

func (c *agentClient) sendAsyncTaskMessage(method string, arguments []interface{}) (value map[string]interface{}, err error) {
	var response TaskResponse
	err = c.agentRequest.Send(method, arguments, &response)
	if err != nil {
		return value, bosherr.WrapErrorf(err, "Sending '%s' to the agent", method)
	}

	agentTaskID, err := response.TaskID()
	if err != nil {
		return value, bosherr.WrapError(err, "Getting agent task id")
	}

	getTaskRetryable := boshretry.NewRetryable(func() (bool, error) {
		var response TaskResponse
		err = c.agentRequest.Send("get_task", []interface{}{agentTaskID}, &response)
		if err != nil {
			return false, bosherr.WrapError(err, "Sending 'get_task' to the agent")
		}

		c.logger.Debug(c.logTag, "get_task response value: %#v", response.Value)

		taskState, err := response.TaskState()
		if err != nil {
			return false, bosherr.WrapError(err, "Getting task state")
		}

		if taskState != "running" {
			var ok bool
			value, ok = response.Value.(map[string]interface{})
			if !ok {
				c.logger.Warn(c.logTag, "Unable to parse get_task response value: %#v", response.Value)
			}
			return true, nil
		}

		return true, bosherr.Errorf("Task %s is still running", method)
	})

	getTaskRetryStrategy := boshretry.NewUnlimitedRetryStrategy(c.getTaskDelay, getTaskRetryable, c.logger)
	return value, getTaskRetryStrategy.Try()
}

func (c *agentClient) CompilePackage(packageSource bmagentclient.BlobRef, compiledPackageDependencies []bmagentclient.BlobRef) (compiledPackageRef bmagentclient.BlobRef, err error) {
	dependencies := make(map[string]BlobRef, len(compiledPackageDependencies))
	for _, dependency := range compiledPackageDependencies {
		dependencies[dependency.Name] = BlobRef{
			Name:        dependency.Name,
			Version:     dependency.Version,
			SHA1:        dependency.SHA1,
			BlobstoreID: dependency.BlobstoreID,
		}
	}

	args := []interface{}{
		packageSource.BlobstoreID,
		packageSource.SHA1,
		packageSource.Name,
		packageSource.Version,
		dependencies,
	}

	responseValue, err := c.sendAsyncTaskMessage("compile_package", args)
	if err != nil {
		return bmagentclient.BlobRef{}, bosherr.WrapError(err, "Sending 'compile_package' to the agent")
	}

	result, ok := responseValue["result"].(map[string]interface{})
	if !ok {
		return bmagentclient.BlobRef{}, bosherr.Errorf("Unable to parse 'compile_package' response from the agent: %#v", responseValue)
	}

	sha1, ok := result["sha1"].(string)
	if !ok {
		return bmagentclient.BlobRef{}, bosherr.Errorf("Unable to parse 'compile_package' response from the agent: %#v", responseValue)
	}

	blobstoreID, ok := result["blobstore_id"].(string)
	if !ok {
		return bmagentclient.BlobRef{}, bosherr.Errorf("Unable to parse 'compile_package' response from the agent: %#v", responseValue)
	}

	compiledPackageRef = bmagentclient.BlobRef{
		Name:        packageSource.Name,
		Version:     packageSource.Version,
		SHA1:        sha1,
		BlobstoreID: blobstoreID,
	}

	return compiledPackageRef, nil
}
