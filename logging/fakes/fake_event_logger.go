package fakes

import (
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
)

type FakeEventLogger struct {
	LoggedEvents   []bmlog.Event
	AddEventErrors map[bmlog.EventState]error
}

func NewFakeEventLogger() *FakeEventLogger {
	return &FakeEventLogger{
		AddEventErrors: map[bmlog.EventState]error{},
	}
}

func (fl *FakeEventLogger) AddEvent(event bmlog.Event) error {
	fl.LoggedEvents = append(fl.LoggedEvents, event)
	err, found := fl.AddEventErrors[event.State]
	if found {
		return err
	}
	return nil
}
