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

const (
	startedState  = "started"
	finishedState = "finished"
	failedState   = "failed"
)

func (c releaseCompiler) Compile(release bmrel.Release) error {
	packages, err := c.dependencyAnalysis.DeterminePackageCompilationOrder(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release")
	}

	stage := "compiling packages"
	totalCount := len(packages)

	for index, pkg := range packages {
		task := fmt.Sprintf("%s/%s", pkg.Name, pkg.Fingerprint)
		startEvent := bmlog.Event{
			Time:  c.timeService.Now(),
			Stage: stage,
			Total: totalCount,
			State: startedState,
			Index: index + 1,
			Task:  task,
		}
		c.eventLogger.AddEvent(startEvent)
		err = c.packageCompiler.Compile(pkg)

		if err != nil {
			failEvent := bmlog.Event{
				Time:  c.timeService.Now(),
				Stage: stage,
				Total: totalCount,
				State: failedState,
				Index: index + 1,
				Task:  task,
			}
			c.eventLogger.AddEvent(failEvent)
			return bosherr.WrapError(err, fmt.Sprintf("Package `%s' compilation failed", pkg.Name))
		}
		stopEvent := bmlog.Event{
			Time:  c.timeService.Now(),
			Stage: stage,
			Total: totalCount,
			State: finishedState,
			Index: index + 1,
			Task:  task,
		}
		c.eventLogger.AddEvent(stopEvent)
	}

	return nil
}
