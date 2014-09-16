package cloud

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type Cloud interface {
	bmstemcell.Infrastructure
}

type cloud struct {
}

func (c cloud) CreateStemcell(bmstemcell.Stemcell) (bmstemcell.CID, error) {
	return bmstemcell.CID(""), nil
}

func NewCloud() Cloud {
	return cloud{}
}
