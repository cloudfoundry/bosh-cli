package compile

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type ReleasePackagesCompiler interface {
	Compile(bmrel.Release) error
}

type releasePackagesCompiler struct {
	dependencyAnalysis DependencyAnalysis
	packageCompiler    PackageCompiler
	eventLogger        bmeventlog.EventLogger
	timeService        boshtime.Service
}

func NewReleasePackagesCompiler(
	da DependencyAnalysis,
	packageCompiler PackageCompiler,
	eventLogger bmeventlog.EventLogger,
	timeService boshtime.Service,
) ReleasePackagesCompiler {
	return &releasePackagesCompiler{
		dependencyAnalysis: da,
		packageCompiler:    packageCompiler,
		eventLogger:        eventLogger,
		timeService:        timeService,
	}
}

func (c releasePackagesCompiler) Compile(release bmrel.Release) error {
	packages, err := c.dependencyAnalysis.DeterminePackageCompilationOrder(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release")
	}

	totalCount := len(packages)
	for index, pkg := range packages {
		logErr := c.compilationEvent(totalCount, index+1, pkg, bmeventlog.Started, "")
		if logErr != nil {
			return logErr
		}

		err = c.packageCompiler.Compile(pkg)

		if err != nil {
			logErr := c.compilationEvent(totalCount, index+1, pkg, bmeventlog.Failed, err.Error())
			if logErr != nil {
				return logErr
			}

			return bosherr.WrapError(err, fmt.Sprintf("Package `%s' compilation failed", pkg.Name))
		}
		logErr = c.compilationEvent(totalCount, index+1, pkg, bmeventlog.Finished, "")
		if logErr != nil {
			return logErr
		}
	}

	return nil
}

func (c releasePackagesCompiler) compilationEvent(
	totalCount,
	index int,
	pkg *bmrel.Package,
	state bmeventlog.EventState,
	message string,
) error {
	stage := "compiling packages"
	task := fmt.Sprintf("%s/%s", pkg.Name, pkg.Fingerprint)
	event := bmeventlog.Event{
		Time:    c.timeService.Now(),
		Stage:   stage,
		Total:   totalCount,
		State:   state,
		Index:   index,
		Task:    task,
		Message: message,
	}
	logErr := c.eventLogger.AddEvent(event)
	if logErr != nil {
		return bosherr.WrapError(logErr, "Logging event: %#v", event)
	}
	return nil
}
