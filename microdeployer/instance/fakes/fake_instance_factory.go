package fakes

import (
	bmins "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instance"
)

type FakeInstanceFactory struct {
	CreateMbusURL  string
	CreateInstance bmins.Instance
}

func NewFakeInstanceFactory() *FakeInstanceFactory {
	return &FakeInstanceFactory{}
}

func (f *FakeInstanceFactory) Create(mbusURL string) bmins.Instance {
	f.CreateMbusURL = mbusURL
	return f.CreateInstance
}
