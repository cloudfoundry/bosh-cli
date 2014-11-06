package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeStep struct {
	Name        string
	States      []bmeventlog.EventState
	SkipMessage string
	FailMessage string
}

func (s *FakeStep) Start() {
	s.States = append(s.States, bmeventlog.Started)
}

func (s *FakeStep) Finish() {
	s.States = append(s.States, bmeventlog.Finished)
}

func (s *FakeStep) Skip(message string) {
	s.States = append(s.States, bmeventlog.Skipped)
	s.SkipMessage = message
}

func (s *FakeStep) Fail(message string) {
	s.States = append(s.States, bmeventlog.Failed)
	s.FailMessage = message
}
