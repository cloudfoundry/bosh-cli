package pkg

import (
	birelpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Compiler

type Compiler interface {
	Compile(birelpkg.Compilable) (CompiledPackageRecord, bool, error)
}
