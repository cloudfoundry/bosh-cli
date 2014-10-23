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
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
)

type AgentClient interface {
	Ping() (string, error)
	Stop() error
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
	Arguments []string
	ReplyTo   string `json:"reply_to"`
}

type agentResponse interface {
	GetException() exceptionResponse
}

type agentSimpleResponse struct {
	Value     string
	Exception exceptionResponse
}

func (r *agentSimpleResponse) GetException() exceptionResponse {
	return r.Exception
}

type agentTaskResponse struct {
	Value     taskState
	Exception exceptionResponse
}

func (r *agentTaskResponse) GetException() exceptionResponse {
	return r.Exception
}

type taskState struct {
	AgentTaskID string `json:"agent_task_id"`
	State       string
}

type exceptionResponse struct {
	Message string
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
	responseBody, err := c.sendMessage("ping", []string{})
	if err != nil {
		return "", bosherr.WrapError(err, "Sending ping message to agent")
	}

	var response agentSimpleResponse
	err = c.handleResponse(responseBody, &response)
	if err != nil {
		return "", bosherr.WrapError(err, "Handling agent response")
	}

	return response.Value, nil
}

func (c *agentClient) Stop() error {
	responseBody, err := c.sendMessage("stop", []string{})
	if err != nil {
		return bosherr.WrapError(err, "Sending stop message to agent")
	}

	var response agentTaskResponse
	err = c.handleResponse(responseBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Handling agent response")
	}

	getTaskRetryable := bmretrystrategy.NewRetryable(func() error {
		responseBody, err := c.sendMessage("get_task", []string{response.Value.AgentTaskID})
		var response agentTaskResponse
		err = c.handleResponse(responseBody, &response)
		if err != nil {
			return bosherr.WrapError(err, "Handling agent response")
		}

		if response.Value.State != "running" {
			return nil
		}

		return bosherr.New("Stop task is still running")
	})

	getTaskRetryStrategy := bmretrystrategy.NewAttemptRetryStrategy(600, c.getTaskDelay, getTaskRetryable, c.logger)
	return getTaskRetryStrategy.Try()
}

func (c *agentClient) sendMessage(method string, arguments []string) ([]byte, error) {
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

func (c *agentClient) handleResponse(responseBody []byte, response agentResponse) error {
	err := json.Unmarshal(responseBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling agent response")
	}

	if (response.GetException() != exceptionResponse{}) {
		return bosherr.New("Agent responded with error: %s", response.GetException().Message)
	}

	return nil
}
