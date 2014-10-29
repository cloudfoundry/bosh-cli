package agentclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmas "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
)

type AgentClient interface {
	Ping() (string, error)
	Stop() error
	Apply(bmas.ApplySpec) error
	Start() error
}

type agentClient struct {
	endpoint     string
	uuid         string
	httpClient   http.Client
	getTaskDelay time.Duration
	logger       boshlog.Logger
	logTag       string
}

type AgentRequest struct {
	Method    string
	Arguments []interface{}
	ReplyTo   string `json:"reply_to"`
}

func NewAgentClient(endpoint string, uuid string, getTaskDelay time.Duration, logger boshlog.Logger) AgentClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := http.Client{Transport: tr}

	return &agentClient{
		endpoint:     fmt.Sprintf("%s/agent", endpoint),
		uuid:         uuid,
		httpClient:   httpClient,
		getTaskDelay: getTaskDelay,
		logger:       logger,
		logTag:       "agentClient",
	}
}

func (c *agentClient) Ping() (string, error) {
	return c.sendSyncMessage("ping", []interface{}{})
}

func (c *agentClient) Stop() error {
	return c.sendAsyncMessage("stop", []interface{}{})
}

func (c *agentClient) Apply(spec bmas.ApplySpec) error {
	return c.sendAsyncMessage("apply", []interface{}{spec})
}

func (c *agentClient) Start() error {
	response, err := c.sendSyncMessage("start", []interface{}{})
	if err != nil {
		return bosherr.WrapError(err, "Starting agent services")
	}

	if response != "started" {
		return bosherr.New("Failed to start agent services with response: '%s'", response)
	}

	return nil
}

func (c *agentClient) sendSyncMessage(method string, arguments []interface{}) (string, error) {
	responseBody, err := c.sendMessage(method, []interface{}{})
	if err != nil {
		return "", bosherr.WrapError(err, "Sending '%s' message to agent", method)
	}

	var response SimpleResponse
	err = c.handleResponse(responseBody, &response)
	if err != nil {
		return "", bosherr.WrapError(err, "Handling agent response")
	}

	return response.Value, nil
}

func (c *agentClient) sendAsyncMessage(method string, arguments []interface{}) error {
	responseBody, err := c.sendMessage(method, arguments)
	if err != nil {
		return bosherr.WrapError(err, "Sending %s message to agent", method)
	}

	var response TaskResponse
	err = c.handleResponse(responseBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Handling agent response")
	}

	agentTaskID, err := response.TaskID()
	if err != nil {
		return bosherr.WrapError(err, "Getting agent task id")
	}

	getTaskRetryable := bmretrystrategy.NewRetryable(func() (bool, error) {
		responseBody, err := c.sendMessage("get_task", []interface{}{agentTaskID})
		var response TaskResponse
		err = c.handleResponse(responseBody, &response)
		if err != nil {
			return false, bosherr.WrapError(err, "Handling agent response")
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

func (c *agentClient) sendMessage(method string, arguments []interface{}) ([]byte, error) {
	postBody := AgentRequest{
		Method:    method,
		Arguments: arguments,
		ReplyTo:   c.uuid,
	}

	httpResponse, err := c.doPost(c.endpoint, postBody)
	if err != nil {
		return []byte{}, bosherr.WrapError(err, "Sending %s to agent", method)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return []byte{}, bosherr.New("Agent responded with non-successful status code: %d", httpResponse.StatusCode)
	}

	httpBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return []byte{}, bosherr.WrapError(err, "Reading agent response")
	}

	return httpBody, nil
}

func (c *agentClient) doPost(endpoint string, agentRequest AgentRequest) (*http.Response, error) {
	agentRequestJSON, err := json.Marshal(agentRequest)
	if err != nil {
		return &http.Response{}, bosherr.WrapError(err, "Marshaling agent request")
	}
	postPayload := strings.NewReader(string(agentRequestJSON))

	c.logger.Debug(c.logTag, "Sending POST request with body %s, endpoint %s", agentRequestJSON, endpoint)

	request, err := http.NewRequest("POST", endpoint, postPayload)
	if err != nil {
		return &http.Response{}, bosherr.WrapError(err, "Creating POST request")
	}

	httpResponse, err := c.httpClient.Do(request)
	if err != nil {
		return &http.Response{}, bosherr.WrapError(err, "Performing POST request")
	}

	return httpResponse, nil
}

func (c *agentClient) handleResponse(responseBody []byte, response Response) error {
	err := json.Unmarshal(responseBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling agent response")
	}

	if (response.GetException() != exceptionResponse{}) {
		return bosherr.New("Agent responded with error: %s", response.GetException().Message)
	}

	return nil
}
