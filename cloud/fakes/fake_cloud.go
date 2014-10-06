package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"

	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"
)

type FakeCloud struct {
	Infrastructure *fakebmstemcell.FakeInfrastructure
}

func NewFakeCloud() *FakeCloud {
	return &FakeCloud{
		Infrastructure: fakebmstemcell.NewFakeInfrastructure(),
	}
}

func (c *FakeCloud) CreateStemcell(stemcell bmstemcell.Stemcell) (bmstemcell.CID, error) {
	return c.Infrastructure.CreateStemcell(stemcell)
}

func (c *FakeCloud) CreateVM(
	cid bmstemcell.CID,
	resourcePoolsSpec map[string]interface{},
	networksSpec map[string]interface{},
) (bmvm.CID, error) {
	return "", nil
}
