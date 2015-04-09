package pkg

import (
	bmrelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
)

type Compiler interface {
	Compile(*bmrelpkg.Package) (CompiledPackageRecord, error)
}
