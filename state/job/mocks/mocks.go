// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudfoundry/bosh-cli/v7/state/job (interfaces: DependencyCompiler)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	job "github.com/cloudfoundry/bosh-cli/v7/release/job"
	job0 "github.com/cloudfoundry/bosh-cli/v7/state/job"
	ui "github.com/cloudfoundry/bosh-cli/v7/ui"
	gomock "github.com/golang/mock/gomock"
)

// MockDependencyCompiler is a mock of DependencyCompiler interface.
type MockDependencyCompiler struct {
	ctrl     *gomock.Controller
	recorder *MockDependencyCompilerMockRecorder
}

// MockDependencyCompilerMockRecorder is the mock recorder for MockDependencyCompiler.
type MockDependencyCompilerMockRecorder struct {
	mock *MockDependencyCompiler
}

// NewMockDependencyCompiler creates a new mock instance.
func NewMockDependencyCompiler(ctrl *gomock.Controller) *MockDependencyCompiler {
	mock := &MockDependencyCompiler{ctrl: ctrl}
	mock.recorder = &MockDependencyCompilerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDependencyCompiler) EXPECT() *MockDependencyCompilerMockRecorder {
	return m.recorder
}

// Compile mocks base method.
func (m *MockDependencyCompiler) Compile(arg0 []job.Job, arg1 ui.Stage) ([]job0.CompiledPackageRef, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Compile", arg0, arg1)
	ret0, _ := ret[0].([]job0.CompiledPackageRef)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Compile indicates an expected call of Compile.
func (mr *MockDependencyCompilerMockRecorder) Compile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Compile", reflect.TypeOf((*MockDependencyCompiler)(nil).Compile), arg0, arg1)
}
