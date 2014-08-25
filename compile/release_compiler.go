package compile

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type ReleaseCompiler interface {
	Compile(bmrel.Release) error
}

type releaseCompiler struct {
	dependencyAnalysis DependencyAnalysis
	packageCompiler    PackageCompiler
	eventLogger        bmlog.EventLogger
}

func NewReleaseCompiler(
	da DependencyAnalysis,
	packageCompiler PackageCompiler,
	eventLogger bmlog.EventLogger,
) ReleaseCompiler {
	return &releaseCompiler{
		dependencyAnalysis: da,
		packageCompiler:    packageCompiler,
		eventLogger:        eventLogger,
	}
}

func (c releaseCompiler) Compile(release bmrel.Release) error {
	packages, err := c.dependencyAnalysis.DeterminePackageCompilationOrder(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release")
	}

	c.eventLogger.StartGroup("compiling packages")
	for _, pkg := range packages {
		err := c.eventLogger.TrackAndLog(fmt.Sprintf("%s/%s", pkg.Name, pkg.Fingerprint), func() error {
			return c.packageCompiler.Compile(pkg)
		})

		if err != nil {
			return bosherr.WrapError(err, fmt.Sprintf("Package `%s' compilation failed", pkg.Name))
		}
	}
	c.eventLogger.FinishGroup()

	return nil
}
