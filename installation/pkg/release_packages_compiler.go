package pkg

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type ReleasePackagesCompiler interface {
	Compile(bmrel.Release) error
}

type releasePackagesCompiler struct {
	packageCompiler PackageCompiler
	eventLogger     bmeventlog.EventLogger
	timeService     boshtime.Service
}

func NewReleasePackagesCompiler(
	packageCompiler PackageCompiler,
	eventLogger bmeventlog.EventLogger,
	timeService boshtime.Service,
) ReleasePackagesCompiler {
	return &releasePackagesCompiler{
		packageCompiler: packageCompiler,
		eventLogger:     eventLogger,
		timeService:     timeService,
	}
}

func (c releasePackagesCompiler) Compile(release bmrel.Release) error {
	eventLoggerStage := c.eventLogger.NewStage("compiling packages")
	eventLoggerStage.Start()
	defer eventLoggerStage.Finish()

	packages := Sort(release.Packages())

	for _, pkg := range packages {
		stepName := fmt.Sprintf("%s/%s", pkg.Name, pkg.Fingerprint)
		err := eventLoggerStage.PerformStep(stepName, func() error {
			err := c.packageCompiler.Compile(pkg)

			if err != nil {
				return bosherr.WrapError(err, fmt.Sprintf("Package '%s' compilation failed", pkg.Name))
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
