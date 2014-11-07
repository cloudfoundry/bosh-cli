package fakes

import (
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
)

type FakeAgentClient struct {
	PingResponses   []pingResponse
	PingCalledCount int

	StopCalled bool
	stopErr    error

	ApplyApplySpec bmas.ApplySpec
	ApplyErr       error

	StartCalled bool
	startErr    error

	MountDiskCID string
	mountDiskErr error

	listDiskDisks  []string
	listDiskErr    error
	ListDiskCalled bool

	GetStateCalledTimes int
	getStateOutputs     []getStateOutput
}

type pingResponse struct {
	response string
	err      error
}

type getStateOutput struct {
	state bmagentclient.State
	err   error
}

func NewFakeAgentClient() *FakeAgentClient {
	return &FakeAgentClient{
		getStateOutputs: []getStateOutput{},
	}
}

func (c *FakeAgentClient) Ping() (string, error) {
	c.PingCalledCount++

	if len(c.PingResponses) > 0 {
		response := c.PingResponses[0]
		c.PingResponses = c.PingResponses[1:]
		return response.response, response.err
	}

	return "", nil
}

func (c *FakeAgentClient) Stop() error {
	c.StopCalled = true
	return c.stopErr
}

func (c *FakeAgentClient) Apply(applySpec bmas.ApplySpec) error {
	c.ApplyApplySpec = applySpec

	return c.ApplyErr
}

func (c *FakeAgentClient) Start() error {
	c.StartCalled = true
	return c.startErr
}

func (c *FakeAgentClient) GetState() (bmagentclient.State, error) {
	c.GetStateCalledTimes++

	getStateReturn := c.getStateOutputs[0]
	c.getStateOutputs = c.getStateOutputs[1:]

	return getStateReturn.state, getStateReturn.err
}

func (c *FakeAgentClient) ListDisk() ([]string, error) {
	c.ListDiskCalled = true
	return c.listDiskDisks, c.listDiskErr
}

func (c *FakeAgentClient) MountDisk(diskCID string) error {
	c.MountDiskCID = diskCID

	return c.mountDiskErr
}

func (c *FakeAgentClient) SetPingBehavior(response string, err error) {
	c.PingResponses = append(c.PingResponses, pingResponse{
		response: response,
		err:      err,
	})
}

func (c *FakeAgentClient) SetStopBehavior(err error) {
	c.stopErr = err
}

func (c *FakeAgentClient) SetStartBehavior(err error) {
	c.startErr = err
}

func (c *FakeAgentClient) SetGetStateBehavior(stateResponse bmagentclient.State, err error) {
	c.getStateOutputs = append(c.getStateOutputs, getStateOutput{
		state: stateResponse,
		err:   err,
	})
}

func (c *FakeAgentClient) SetMountDiskBehavior(err error) {
	c.mountDiskErr = err
}

func (c *FakeAgentClient) SetListDiskBehavior(disks []string, err error) {
	c.listDiskDisks = disks
	c.listDiskErr = err
}
