package http

import (
	"fmt"
	"time"

	biagentclient "github.com/cloudfoundry/bosh-init/deployment/agentclient"
	bias "github.com/cloudfoundry/bosh-init/deployment/applyspec"
	bihttpclient "github.com/cloudfoundry/bosh-init/deployment/httpclient"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
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
	httpClient bihttpclient.HTTPClient,
	logger boshlog.Logger,
) biagentclient.AgentClient {
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

func (c *agentClient) Apply(spec bias.ApplySpec) error {
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

func (c *agentClient) GetState() (biagentclient.AgentState, error) {
	var response StateResponse
	err := c.agentRequest.Send("get_state", []interface{}{}, &response)
	if err != nil {
		return biagentclient.AgentState{}, bosherr.WrapError(err, "Sending get_state to the agent")
	}

	agentState := biagentclient.AgentState{
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
	// cannot call getTaskRetryStrategy.Try in the return statement due to gccgo
	// execution order issues: https://code.google.com/p/go/issues/detail?id=8698&thanks=8698&ts=1410376474
	err = getTaskRetryStrategy.Try()
	return value, err
}

func (c *agentClient) CompilePackage(packageSource biagentclient.BlobRef, compiledPackageDependencies []biagentclient.BlobRef) (compiledPackageRef biagentclient.BlobRef, err error) {
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
		return biagentclient.BlobRef{}, bosherr.WrapError(err, "Sending 'compile_package' to the agent")
	}

	result, ok := responseValue["result"].(map[string]interface{})
	if !ok {
		return biagentclient.BlobRef{}, bosherr.Errorf("Unable to parse 'compile_package' response from the agent: %#v", responseValue)
	}

	sha1, ok := result["sha1"].(string)
	if !ok {
		return biagentclient.BlobRef{}, bosherr.Errorf("Unable to parse 'compile_package' response from the agent: %#v", responseValue)
	}

	blobstoreID, ok := result["blobstore_id"].(string)
	if !ok {
		return biagentclient.BlobRef{}, bosherr.Errorf("Unable to parse 'compile_package' response from the agent: %#v", responseValue)
	}

	compiledPackageRef = biagentclient.BlobRef{
		Name:        packageSource.Name,
		Version:     packageSource.Version,
		SHA1:        sha1,
		BlobstoreID: blobstoreID,
	}

	return compiledPackageRef, nil
}
