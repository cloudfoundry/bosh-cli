package compile

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type ReleaseCompiler interface {
	Compile(bmrel.Release) error
}

type releaseCompiler struct {
	dependencyAnalysis DependencyAnalysis
	packageCompiler    PackageCompiler
}

func NewReleaseCompiler(da DependencyAnalysis, packageCompiler PackageCompiler) ReleaseCompiler {
	return &releaseCompiler{
		dependencyAnalysis: da,
		packageCompiler:    packageCompiler,
	}
}

func (c releaseCompiler) Compile(release bmrel.Release) error {
	packages, err := c.dependencyAnalysis.DeterminePackageCompilationOrder(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release")
	}

	for _, pkg := range packages {
		err := c.packageCompiler.Compile(pkg)
		if err != nil {
			return bosherr.WrapError(err, fmt.Sprintf("Package `%s' compilation failed", pkg.Name))
		}
	}

	return nil
}
