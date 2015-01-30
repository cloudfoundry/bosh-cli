package fakes

import (
	"fmt"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type FakeEventLogger struct {
	LoggedEvents   []bmeventlog.Event
	AddEventErrors map[bmeventlog.EventState]error

	NewStageInputs     []NewStageInput
	stageStageBehavior map[string]bmeventlog.Stage

	StartStageInputs  []string
	FinishStageInputs []string
}

type NewStageInput struct {
	Name string
}

func NewFakeEventLogger() *FakeEventLogger {
	return &FakeEventLogger{
		AddEventErrors:     map[bmeventlog.EventState]error{},
		stageStageBehavior: map[string]bmeventlog.Stage{},
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

	stage, found := fl.stageStageBehavior[name]
	if !found {
		panic(fmt.Sprintf("Recieved unexpected NewStage('%s')", name))
	}
	return stage
}

func (fl *FakeEventLogger) StartStage(name string) {
	fl.StartStageInputs = append(fl.StartStageInputs, name)
}

func (fl *FakeEventLogger) FinishStage(name string) {
	fl.FinishStageInputs = append(fl.FinishStageInputs, name)
}

func (fl *FakeEventLogger) SetNewStageBehavior(name string, stage bmeventlog.Stage) {
	fl.stageStageBehavior[name] = stage
}
