package fakes

import (
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
)

type FakeFactory struct {
	SSHTunnel           bmsshtunnel.SSHTunnel
	NewSSHTunnelOptions bmsshtunnel.Options
}

func NewFakeFactory() *FakeFactory {
	return &FakeFactory{}
}

func (f *FakeFactory) NewSSHTunnel(options bmsshtunnel.Options) bmsshtunnel.SSHTunnel {
	f.NewSSHTunnelOptions = options

	return f.SSHTunnel
}
