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
}

func NewEventLogger(ui bmui.UI, timeService boshtime.Service) EventLogger {
	return eventLogger{
		ui:          ui,
		timeService: timeService,
	}
}

type eventLogger struct {
	ui          bmui.UI
	timeService boshtime.Service
}

func (e eventLogger) TrackAndLog(event string, f func() error) error {
	if event == "" {
		return bosherr.New("TrackAndLog given an empty string as event")
	}

	e.ui.Say(fmt.Sprintf("Started %s.", event))
	startedTime := e.timeService.Now()

	err := f()
	if err != nil {
		return bosherr.WrapError(err, "Event returned error")
	}

	endTime := e.timeService.Now()
	e.ui.Say(fmt.Sprintf(" Done (%s)", bmdfmt.Format(endTime.Sub(startedTime))))

	return nil
}
