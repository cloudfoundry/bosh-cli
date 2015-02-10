package pkg

import (
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type Compiler interface {
	Compile(*bmrelpkg.Package) (CompiledPackageRecord, error)
}
