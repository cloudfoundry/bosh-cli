package agentclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type AgentClient interface {
	Ping() (string, error)
}

type agentClient struct {
	endpoint   string
	uuid       string
	httpClient http.Client
	logger     boshlog.Logger
	logTag     string
}

type agentRequest struct {
	Method    string
	Arguments []string
	ReplyTo   string `json:"reply_to"`
}

type agentResponse struct {
	Value     string
	Exception exceptionResponse
}

type exceptionResponse struct {
	Message string
}

func NewAgentClient(endpoint string, uuid string, logger boshlog.Logger) AgentClient {
	return &agentClient{
		endpoint:   fmt.Sprintf("%s/agent", endpoint),
		uuid:       uuid,
		httpClient: http.Client{},
		logger:     logger,
		logTag:     "agentClient",
	}
}

func (c *agentClient) Ping() (string, error) {
	postBody := agentRequest{
		Method:  "ping",
		ReplyTo: c.uuid,
	}

	httpResponse, err := c.doPost(c.endpoint, postBody)
	if err != nil {
		return "", bosherr.WrapError(err, "Sending ping to agent")
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return "", bosherr.New("Agent responded with non-successful status code: %d", httpResponse.StatusCode)
	}

	httpBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return "", bosherr.WrapError(err, "Reading agent response")
	}

	var response agentResponse
	err = json.Unmarshal(httpBody, &response)
	if err != nil {
		return "", bosherr.WrapError(err, "Unmarshaling agent response")
	}

	if response.Value != "" {
		return response.Value, nil
	}

	return "", bosherr.New("Agent responded with error: %s", response.Exception.Message)
}

func (c *agentClient) doPost(endpoint string, agentRequest agentRequest) (*http.Response, error) {
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
