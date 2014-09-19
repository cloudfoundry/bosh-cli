package fakes

import (
	"fmt"
)

type FakeDevicePathResolver struct {
	realDevicePaths           map[string]string
	GetRealDevicePathTimedOut bool
	GetRealDevicePathErr      error
}

func NewFakeDevicePathResolver() *FakeDevicePathResolver {
	return &FakeDevicePathResolver{realDevicePaths: map[string]string{}}
}

func (r *FakeDevicePathResolver) RegisterRealDevicePath(devicePath, realDevicePath string) {
	_, found := r.realDevicePaths[devicePath]
	if found {
		panic(fmt.Sprintf("Already registered %s", devicePath))
	}
	r.realDevicePaths[devicePath] = realDevicePath
}

func (r *FakeDevicePathResolver) GetRealDevicePath(devicePath string) (string, bool, error) {
	if r.GetRealDevicePathErr != nil {
		return "", r.GetRealDevicePathTimedOut, r.GetRealDevicePathErr
	}

	realDevicePath, found := r.realDevicePaths[devicePath]
	if !found {
		panic(fmt.Sprintf("Could not find real device path for %s", devicePath))
	}
	return realDevicePath, false, nil
}
