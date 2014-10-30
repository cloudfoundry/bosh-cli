package fakes

import (
	bminsup "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instanceupdater"
)

type FakeInstanceFactory struct {
	CreateMbusURL  string
	CreateInstance bminsup.Instance
}

func NewFakeInstanceFactory() *FakeInstanceFactory {
	return &FakeInstanceFactory{}
}

func (f *FakeInstanceFactory) Create(mbusURL string) bminsup.Instance {
	f.CreateMbusURL = mbusURL
	return f.CreateInstance
}
