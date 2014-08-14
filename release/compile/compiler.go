package compile

import (
	boshcomp "github.com/cloudfoundry/bosh-agent/agent/compiler"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
)

type Compiler interface {
	Compile(bmrelease.Release) error
}

type compiler struct {
	dependencyAnalysis DependencyAnalysis
}

func NewCompiler(da DependencyAnalysis, packageCompiler boshcomp.Compiler) Compiler {
	return &compiler{
		dependencyAnalysis: da,
	}
}

func (c compiler) Compile(release bmrelease.Release) error {
	_, err := c.dependencyAnalysis.DeterminePackageCompilationOrder(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release")
	}

	return nil
}
