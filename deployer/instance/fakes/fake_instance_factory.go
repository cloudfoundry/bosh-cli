package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmins "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
)

type FakeInstanceFactory struct {
	CreateInput    CreateInstanceInput
	CreateInstance bmins.Instance
}

type CreateInstanceInput struct {
	VMCID   string
	MbusURL string
	Cloud   bmcloud.Cloud
}

func NewFakeInstanceFactory() *FakeInstanceFactory {
	return &FakeInstanceFactory{}
}

func (f *FakeInstanceFactory) Create(vmCID string, mbusURL string, cloud bmcloud.Cloud) bmins.Instance {
	f.CreateInput = CreateInstanceInput{
		VMCID:   vmCID,
		MbusURL: mbusURL,
		Cloud:   cloud,
	}
	return f.CreateInstance
}
