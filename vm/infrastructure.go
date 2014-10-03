package vm

import (
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

type Infrastructure interface {
	CreateVM(bmstemcell.CID, map[string]interface{}) (CID, error)
}
