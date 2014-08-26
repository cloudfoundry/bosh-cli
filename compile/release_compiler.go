package compile

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

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
	timeService        boshtime.Service
}

func NewReleaseCompiler(
	da DependencyAnalysis,
	packageCompiler PackageCompiler,
	eventLogger bmlog.EventLogger,
	timeService boshtime.Service,
) ReleaseCompiler {
	return &releaseCompiler{
		dependencyAnalysis: da,
		packageCompiler:    packageCompiler,
		eventLogger:        eventLogger,
		timeService:        timeService,
	}
}

func (c releaseCompiler) Compile(release bmrel.Release) error {
	packages, err := c.dependencyAnalysis.DeterminePackageCompilationOrder(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release")
	}

	totalCount := len(packages)
	for index, pkg := range packages {
		logErr := c.compilationEvent(totalCount, index+1, pkg, bmlog.Started, "")
		if logErr != nil {
			return logErr
		}

		err = c.packageCompiler.Compile(pkg)

		if err != nil {
			logErr := c.compilationEvent(totalCount, index+1, pkg, bmlog.Failed, err.Error())
			if logErr != nil {
				return logErr
			}

			return bosherr.WrapError(err, fmt.Sprintf("Package `%s' compilation failed", pkg.Name))
		}
		logErr = c.compilationEvent(totalCount, index+1, pkg, bmlog.Finished, "")
		if logErr != nil {
			return logErr
		}
	}

	return nil
}

func (c releaseCompiler) compilationEvent(
	totalCount,
	index int,
	pkg *bmrel.Package,
	state bmlog.EventState,
	message string,
) error {
	stage := "compiling packages"
	task := fmt.Sprintf("%s/%s", pkg.Name, pkg.Fingerprint)
	event := bmlog.Event{
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
