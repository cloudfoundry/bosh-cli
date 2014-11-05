package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeEventLogger struct {
	LoggedEvents   []bmeventlog.Event
	AddEventErrors map[bmeventlog.EventState]error
}

func NewFakeEventLogger() *FakeEventLogger {
	return &FakeEventLogger{
		AddEventErrors: map[bmeventlog.EventState]error{},
	}
}

func (fl *FakeEventLogger) AddEvent(event bmeventlog.Event) error {
	fl.LoggedEvents = append(fl.LoggedEvents, event)
	err, found := fl.AddEventErrors[event.State]
	if found {
		return err
	}
	return nil
}
