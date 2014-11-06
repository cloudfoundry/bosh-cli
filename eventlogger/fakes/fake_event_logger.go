package fakes

import (
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeEventLogger struct {
	LoggedEvents   []bmeventlog.Event
	AddEventErrors map[bmeventlog.EventState]error

	NewStageInputs []NewStageInput
	newStageStage  bmeventlog.Stage
}

type NewStageInput struct {
	Name       string
	TotalSteps int
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

func (fl *FakeEventLogger) NewStage(name string, totalSteps int) bmeventlog.Stage {
	fl.NewStageInputs = append(fl.NewStageInputs, NewStageInput{
		Name:       name,
		TotalSteps: totalSteps,
	})

	return fl.newStageStage
}

func (fl *FakeEventLogger) SetNewStageBehavior(stage bmeventlog.Stage) {
	fl.newStageStage = stage
}
