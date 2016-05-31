package pkg

import (
	birelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
)

type Compiler interface {
	Compile(birelpkg.Compilable) (CompiledPackageRecord, bool, error)
}
