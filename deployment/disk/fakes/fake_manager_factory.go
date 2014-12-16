package fakes

import (
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
)

type NewManagerInput struct {
	Cloud bmcloud.Cloud
}

type FakeManagerFactory struct {
	NewManagerInputs  []NewManagerInput
	NewManagerManager bmdisk.Manager
}

func NewFakeManagerFactory() *FakeManagerFactory {
	return &FakeManagerFactory{
		NewManagerInputs: []NewManagerInput{},
	}
}

func (f *FakeManagerFactory) NewManager(cloud bmcloud.Cloud) bmdisk.Manager {
	input := NewManagerInput{
		Cloud: cloud,
	}
	f.NewManagerInputs = append(f.NewManagerInputs, input)

	return f.NewManagerManager
}
