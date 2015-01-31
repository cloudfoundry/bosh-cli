package pkg

import (
	"fmt"

	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type ReleasePackagesCompiler interface {
	Compile(bmrel.Release, bmeventlog.Stage) error
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

func (c releasePackagesCompiler) Compile(release bmrel.Release, stage bmeventlog.Stage) error {
	//TODO: should just take a list of packages, not a whole release [#85719162]
	// sort release packages in compilation order
	packages := bmrelpkg.Sort(release.Packages())

	for _, pkg := range packages {
		stepName := fmt.Sprintf("Compiling package '%s/%s'", pkg.Name, pkg.Fingerprint)
		err := stage.PerformStep(stepName, func() error {
			return c.packageCompiler.Compile(pkg)
		})
		if err != nil {
			return err
		}
	}

	return nil
}
