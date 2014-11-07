package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeStage struct {
	StageName       string
	StageTotalSteps int
	Steps           []*FakeStep

	Started  bool
	Finished bool
}

func NewFakeStage() *FakeStage {
	return &FakeStage{
		Steps: []*FakeStep{},
	}
}

func (s *FakeStage) NewStep(name string) bmeventlog.Step {
	fakeStep := &FakeStep{
		Name:   name,
		States: []bmeventlog.EventState{},
	}
	s.Steps = append(s.Steps, fakeStep)

	return fakeStep
}

func (s *FakeStage) Name() string {
	return s.StageName
}

func (s *FakeStage) TotalSteps() int {
	return s.StageTotalSteps
}

func (s *FakeStage) Start() {
	s.Started = true
}

func (s *FakeStage) Finish() {
	s.Finished = true
}
