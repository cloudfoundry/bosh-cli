// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudfoundry/bosh-cli/v7/cloud (interfaces: Cloud,Factory)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	cloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	installation "github.com/cloudfoundry/bosh-cli/v7/installation"
	property "github.com/cloudfoundry/bosh-utils/property"
	gomock "github.com/golang/mock/gomock"
)

// MockCloud is a mock of Cloud interface.
type MockCloud struct {
	ctrl     *gomock.Controller
	recorder *MockCloudMockRecorder
}

// MockCloudMockRecorder is the mock recorder for MockCloud.
type MockCloudMockRecorder struct {
	mock *MockCloud
}

// NewMockCloud creates a new mock instance.
func NewMockCloud(ctrl *gomock.Controller) *MockCloud {
	mock := &MockCloud{ctrl: ctrl}
	mock.recorder = &MockCloudMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCloud) EXPECT() *MockCloudMockRecorder {
	return m.recorder
}

// AttachDisk mocks base method.
func (m *MockCloud) AttachDisk(arg0, arg1 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttachDisk", arg0, arg1)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttachDisk indicates an expected call of AttachDisk.
func (mr *MockCloudMockRecorder) AttachDisk(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttachDisk", reflect.TypeOf((*MockCloud)(nil).AttachDisk), arg0, arg1)
}

// CreateDisk mocks base method.
func (m *MockCloud) CreateDisk(arg0 int, arg1 property.Map, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDisk", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDisk indicates an expected call of CreateDisk.
func (mr *MockCloudMockRecorder) CreateDisk(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDisk", reflect.TypeOf((*MockCloud)(nil).CreateDisk), arg0, arg1, arg2)
}

// CreateStemcell mocks base method.
func (m *MockCloud) CreateStemcell(arg0 string, arg1 property.Map) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateStemcell", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateStemcell indicates an expected call of CreateStemcell.
func (mr *MockCloudMockRecorder) CreateStemcell(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateStemcell", reflect.TypeOf((*MockCloud)(nil).CreateStemcell), arg0, arg1)
}

// CreateVM mocks base method.
func (m *MockCloud) CreateVM(arg0, arg1 string, arg2 property.Map, arg3 map[string]property.Map, arg4 property.Map) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateVM", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateVM indicates an expected call of CreateVM.
func (mr *MockCloudMockRecorder) CreateVM(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVM", reflect.TypeOf((*MockCloud)(nil).CreateVM), arg0, arg1, arg2, arg3, arg4)
}

// DeleteDisk mocks base method.
func (m *MockCloud) DeleteDisk(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteDisk", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteDisk indicates an expected call of DeleteDisk.
func (mr *MockCloudMockRecorder) DeleteDisk(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDisk", reflect.TypeOf((*MockCloud)(nil).DeleteDisk), arg0)
}

// DeleteStemcell mocks base method.
func (m *MockCloud) DeleteStemcell(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteStemcell", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteStemcell indicates an expected call of DeleteStemcell.
func (mr *MockCloudMockRecorder) DeleteStemcell(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteStemcell", reflect.TypeOf((*MockCloud)(nil).DeleteStemcell), arg0)
}

// DeleteVM mocks base method.
func (m *MockCloud) DeleteVM(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteVM", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteVM indicates an expected call of DeleteVM.
func (mr *MockCloudMockRecorder) DeleteVM(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteVM", reflect.TypeOf((*MockCloud)(nil).DeleteVM), arg0)
}

// DetachDisk mocks base method.
func (m *MockCloud) DetachDisk(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetachDisk", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DetachDisk indicates an expected call of DetachDisk.
func (mr *MockCloudMockRecorder) DetachDisk(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetachDisk", reflect.TypeOf((*MockCloud)(nil).DetachDisk), arg0, arg1)
}

// HasVM mocks base method.
func (m *MockCloud) HasVM(arg0 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasVM", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HasVM indicates an expected call of HasVM.
func (mr *MockCloudMockRecorder) HasVM(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasVM", reflect.TypeOf((*MockCloud)(nil).HasVM), arg0)
}

// Info mocks base method.
func (m *MockCloud) Info() (cloud.CpiInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(cloud.CpiInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Info indicates an expected call of Info.
func (mr *MockCloudMockRecorder) Info() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockCloud)(nil).Info))
}

// SetDiskMetadata mocks base method.
func (m *MockCloud) SetDiskMetadata(arg0 string, arg1 cloud.DiskMetadata) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDiskMetadata", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDiskMetadata indicates an expected call of SetDiskMetadata.
func (mr *MockCloudMockRecorder) SetDiskMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDiskMetadata", reflect.TypeOf((*MockCloud)(nil).SetDiskMetadata), arg0, arg1)
}

// SetVMMetadata mocks base method.
func (m *MockCloud) SetVMMetadata(arg0 string, arg1 cloud.VMMetadata) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetVMMetadata", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetVMMetadata indicates an expected call of SetVMMetadata.
func (mr *MockCloudMockRecorder) SetVMMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetVMMetadata", reflect.TypeOf((*MockCloud)(nil).SetVMMetadata), arg0, arg1)
}

// String mocks base method.
func (m *MockCloud) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockCloudMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockCloud)(nil).String))
}

// MockFactory is a mock of Factory interface.
type MockFactory struct {
	ctrl     *gomock.Controller
	recorder *MockFactoryMockRecorder
}

// MockFactoryMockRecorder is the mock recorder for MockFactory.
type MockFactoryMockRecorder struct {
	mock *MockFactory
}

// NewMockFactory creates a new mock instance.
func NewMockFactory(ctrl *gomock.Controller) *MockFactory {
	mock := &MockFactory{ctrl: ctrl}
	mock.recorder = &MockFactoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFactory) EXPECT() *MockFactoryMockRecorder {
	return m.recorder
}

// NewCloud mocks base method.
func (m *MockFactory) NewCloud(arg0 installation.Installation, arg1 string, arg2 int) (cloud.Cloud, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewCloud", arg0, arg1, arg2)
	ret0, _ := ret[0].(cloud.Cloud)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewCloud indicates an expected call of NewCloud.
func (mr *MockFactoryMockRecorder) NewCloud(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewCloud", reflect.TypeOf((*MockFactory)(nil).NewCloud), arg0, arg1, arg2)
}
