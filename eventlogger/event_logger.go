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
	Task    string
	State   EventState
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
	NewStage(string) Stage
	AddEvent(event Event) error
	StartStage(string)
	FinishStage(string)
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

func (e *eventLogger) NewStage(name string) Stage {
	return NewStage(name, e)
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
		e.ui.Say(fmt.Sprintf("Started %s", key))
		e.startedTasks[key] = event.Time
	case Finished:
		duration := event.Time.Sub(e.startedTasks[key])
		e.ui.Sayln(fmt.Sprintf(" done. (%s)", durationfmt.Format(duration)))
	case Failed:
		duration := event.Time.Sub(e.startedTasks[key])
		e.ui.Sayln(fmt.Sprintf(" failed (%s). (%s)", event.Message, durationfmt.Format(duration)))
	case Skipped:
		e.ui.Sayln(fmt.Sprintf("Started %s skipped (%s).", key, event.Message))
	default:
		return bosherr.New("Unsupported event state `%s'", event.State)
	}
	return nil
}

func (e *eventLogger) FinishStage(name string) {
	e.ui.Sayln(fmt.Sprintf("Done %s", name))
	e.ui.Sayln("")
}

func (e *eventLogger) StartStage(name string) {
	e.ui.Sayln(fmt.Sprintf("Started %s", name))
}
