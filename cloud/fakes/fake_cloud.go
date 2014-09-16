package fakes

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

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
