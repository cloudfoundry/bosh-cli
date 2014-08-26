package fakes

import (
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
)

type FakeEventLogger struct {
	LoggedEvents []bmlog.Event
}

func NewFakeEventLogger() *FakeEventLogger {
	return &FakeEventLogger{}
}

func (fl *FakeEventLogger) AddEvent(event bmlog.Event) {
	fl.LoggedEvents = append(fl.LoggedEvents, event)
}
