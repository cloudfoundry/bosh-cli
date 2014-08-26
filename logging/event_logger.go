package logging

import (
	"fmt"
	durationfmt "github.com/cloudfoundry/bosh-micro-cli/durationfmt"
	"time"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type EventLogger interface {
	// NEW
	AddEvent(event Event)
}

func NewEventLogger(ui bmui.UI) EventLogger {
	return &eventLogger{
		ui:           ui,
		startedTasks: make(map[string]time.Time),
	}
}

type eventLogger struct {
	ui           bmui.UI
	startedTasks map[string]time.Time
}

func (e *eventLogger) AddEvent(event Event) {
	key := fmt.Sprintf("%s > %s.", event.Stage, event.Task)

	if event.State == "started" {
		if event.Index == 1 {
			e.ui.Sayln(fmt.Sprintf("Started %s", event.Stage))
		}
		e.ui.Say(fmt.Sprintf("Started %s", key))
		e.startedTasks[key] = event.Time
	} else if event.State == "finished" {
		duration := event.Time.Sub(e.startedTasks[key])
		e.ui.Sayln(fmt.Sprintf(" Done (%s)", durationfmt.Format(duration)))
		if event.Index == event.Total {
			e.ui.Sayln(fmt.Sprintf("Done %s", event.Stage))
		}
	}
}
