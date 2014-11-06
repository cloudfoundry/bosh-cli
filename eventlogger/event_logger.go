package eventlogger

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	durationfmt "github.com/cloudfoundry/bosh-micro-cli/eventlogger/durationfmt"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type EventFilter interface {
	Filter(*Event) error
}

type Event struct {
	Time    time.Time
	Stage   string
	Total   int
	Task    string
	State   EventState
	Index   int
	Message string
}

type EventState string

const (
	Started  EventState = "started"
	Finished EventState = "finished"
	Failed   EventState = "failed"
	Skipped  EventState = "skipped"
)

type EventLogger interface {
	NewStage(string, int) Stage
	AddEvent(event Event) error
}

type eventLogger struct {
	ui           bmui.UI
	startedTasks map[string]time.Time
	filters      []EventFilter
}

func NewEventLogger(ui bmui.UI) EventLogger {
	return &eventLogger{
		ui:           ui,
		startedTasks: make(map[string]time.Time),
		filters:      []EventFilter{},
	}
}

func (e *eventLogger) NewStage(name string, totalSteps int) Stage {
	return NewStage(name, totalSteps, e)
}

func NewEventLoggerWithFilters(ui bmui.UI, filters []EventFilter) EventLogger {
	return &eventLogger{
		ui:           ui,
		startedTasks: make(map[string]time.Time),
		filters:      filters,
	}
}

func (e *eventLogger) AddEvent(event Event) error {
	if e.filters != nil && len(e.filters) > 0 {
		for _, filter := range e.filters {
			filter.Filter(&event)
		}
	}

	key := fmt.Sprintf("%s > %s...", event.Stage, event.Task)
	switch event.State {
	case Started:
		if event.Index == 1 {
			e.ui.Sayln(fmt.Sprintf("Started %s", event.Stage))
		}
		e.ui.Say(fmt.Sprintf("Started %s", key))
		e.startedTasks[key] = event.Time
	case Finished:
		duration := event.Time.Sub(e.startedTasks[key])
		e.ui.Sayln(fmt.Sprintf(" done. (%s)", durationfmt.Format(duration)))
		if event.Index == event.Total {
			e.ui.Sayln(fmt.Sprintf("Done %s", event.Stage))
			e.ui.Sayln("")
		}
	case Failed:
		duration := event.Time.Sub(e.startedTasks[key])
		e.ui.Sayln(fmt.Sprintf(" failed (%s). (%s)", event.Message, durationfmt.Format(duration)))
	case Skipped:
		e.ui.Sayln(fmt.Sprintf("Started %s skipped (%s).", key, event.Message))
		if event.Index == event.Total {
			e.ui.Sayln(fmt.Sprintf("Done %s", event.Stage))
			e.ui.Sayln("")
		}
	default:
		return bosherr.New("Unsupported event state `%s'", event.State)
	}
	return nil
}
