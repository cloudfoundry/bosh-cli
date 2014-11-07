package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeEventLogger struct {
	LoggedEvents   []bmeventlog.Event
	AddEventErrors map[bmeventlog.EventState]error

	NewStageInputs []NewStageInput
	newStageStage  bmeventlog.Stage

	StartStageInputs  []string
	FinishStageInputs []string
}

type NewStageInput struct {
	Name string
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

func (fl *FakeEventLogger) NewStage(name string) bmeventlog.Stage {
	fl.NewStageInputs = append(fl.NewStageInputs, NewStageInput{
		Name: name,
	})

	return fl.newStageStage
}

func (fl *FakeEventLogger) StartStage(name string) {
	fl.StartStageInputs = append(fl.StartStageInputs, name)
}

func (fl *FakeEventLogger) FinishStage(name string) {
	fl.FinishStageInputs = append(fl.FinishStageInputs, name)
}

func (fl *FakeEventLogger) SetNewStageBehavior(stage bmeventlog.Stage) {
	fl.newStageStage = stage
}
