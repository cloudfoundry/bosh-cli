package logging

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmdfmt "github.com/cloudfoundry/bosh-micro-cli/durationfmt"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type EventLogger interface {
	TrackAndLog(event string, f func() error) error
	StartGroup(event string) error
	FinishGroup() error
}

func NewEventLogger(ui bmui.UI, timeService boshtime.Service) EventLogger {
	return &eventLogger{
		ui:          ui,
		timeService: timeService,
	}
}

type eventLogger struct {
	ui           bmui.UI
	timeService  boshtime.Service
	startedGroup string
}

func (e eventLogger) TrackAndLog(event string, f func() error) error {
	if event == "" {
		return bosherr.New("TrackAndLog given an empty string as event")
	}

	e.ui.Say(fmt.Sprintf("Started %s > %s", e.startedGroup, event))
	startedTime := e.timeService.Now()

	err := f()
	if err != nil {
		return bosherr.WrapError(err, "Event returned error")
	}

	endTime := e.timeService.Now()
	e.ui.Sayln(fmt.Sprintf(" Done (%s)", bmdfmt.Format(endTime.Sub(startedTime))))

	return nil
}

func (e *eventLogger) StartGroup(group string) error {
	if group == "" {
		return bosherr.New("StartGroup given an empty string as group")
	}
	e.ui.Sayln(fmt.Sprintf("Started %s", group))
	e.startedGroup = group
	return nil
}

func (e *eventLogger) FinishGroup() error {
	if e.startedGroup == "" {
		return bosherr.New("FinishGroup called without a group started")
	}

	e.ui.Sayln(fmt.Sprintf("Done %s", e.startedGroup))
	e.startedGroup = ""
	return nil
}
