package fakes

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type FakeHTTPClient struct {
	PostInputs  []postInput
	postOutputs []postOutput

	GetInputs  []getInput
	getOutputs []getOutput
}

type postInput struct {
	Payload  []byte
	Endpoint string
}

type postOutput struct {
	response *http.Response
	err      error
}

type getInput struct {
	Endpoint string
}

type getOutput struct {
	response *http.Response
	err      error
}

func NewFakeHTTPClient() *FakeHTTPClient {
	return &FakeHTTPClient{
		postOutputs: []postOutput{},
		getOutputs:  []getOutput{},
	}
}

func (c *FakeHTTPClient) Post(endpoint string, payload []byte) (*http.Response, error) {
	c.PostInputs = append(c.PostInputs, postInput{
		Payload:  payload,
		Endpoint: endpoint,
	})

	postReturn := c.postOutputs[0]
	c.postOutputs = c.postOutputs[1:]

	return postReturn.response, postReturn.err
}

func (c *FakeHTTPClient) Get(endpoint string) (*http.Response, error) {
	c.GetInputs = append(c.GetInputs, getInput{
		Endpoint: endpoint,
	})

	getReturn := c.getOutputs[0]
	c.getOutputs = c.getOutputs[1:]

	return getReturn.response, getReturn.err
}

func (c *FakeHTTPClient) SetPostBehavior(body string, statusCode int, err error) {
	postResponse := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		StatusCode: statusCode,
	}
	c.postOutputs = append(c.postOutputs, postOutput{
		response: postResponse,
		err:      err,
	})
}

func (c *FakeHTTPClient) SetGetBehavior(body string, statusCode int, err error) {
	getResponse := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		StatusCode: statusCode,
	}

	c.getOutputs = append(c.getOutputs, getOutput{
		response: getResponse,
		err:      err,
	})
}
