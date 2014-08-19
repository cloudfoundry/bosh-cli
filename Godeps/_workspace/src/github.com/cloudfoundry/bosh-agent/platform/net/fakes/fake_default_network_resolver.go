package fakes

import (
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
)

type FakeDefaultNetworkResolver struct {
	GetDefaultNetworkNetwork boshsettings.Network
	GetDefaultNetworkErr     error
}

func (r *FakeDefaultNetworkResolver) GetDefaultNetwork() (boshsettings.Network, error) {
	return r.GetDefaultNetworkNetwork, r.GetDefaultNetworkErr
}
