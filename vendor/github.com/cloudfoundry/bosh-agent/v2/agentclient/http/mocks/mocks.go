// Code generated by MockGen. DO NOT EDIT.
// Source: agent_client_factory.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	agentclient "github.com/cloudfoundry/bosh-agent/v2/agentclient"
	gomock "github.com/golang/mock/gomock"
)

// MockAgentClientFactory is a mock of AgentClientFactory interface.
type MockAgentClientFactory struct {
	ctrl     *gomock.Controller
	recorder *MockAgentClientFactoryMockRecorder
}

// MockAgentClientFactoryMockRecorder is the mock recorder for MockAgentClientFactory.
type MockAgentClientFactoryMockRecorder struct {
	mock *MockAgentClientFactory
}

// NewMockAgentClientFactory creates a new mock instance.
func NewMockAgentClientFactory(ctrl *gomock.Controller) *MockAgentClientFactory {
	mock := &MockAgentClientFactory{ctrl: ctrl}
	mock.recorder = &MockAgentClientFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAgentClientFactory) EXPECT() *MockAgentClientFactoryMockRecorder {
	return m.recorder
}

// NewAgentClient mocks base method.
func (m *MockAgentClientFactory) NewAgentClient(directorID, mbusURL, caCert string) (agentclient.AgentClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewAgentClient", directorID, mbusURL, caCert)
	ret0, _ := ret[0].(agentclient.AgentClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewAgentClient indicates an expected call of NewAgentClient.
func (mr *MockAgentClientFactoryMockRecorder) NewAgentClient(directorID, mbusURL, caCert interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewAgentClient", reflect.TypeOf((*MockAgentClientFactory)(nil).NewAgentClient), directorID, mbusURL, caCert)
}