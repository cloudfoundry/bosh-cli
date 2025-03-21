package fakes

import (
	"errors"

	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type FakeStage struct {
	PerformCalls []*PerformCall
	SubStages    []*FakeStage
}

type PerformCall struct {
	Name      string
	Error     error
	SkipError error
	Stage     *FakeStage
}

func NewFakeStage() *FakeStage {
	return &FakeStage{}
}

func (s *FakeStage) Perform(name string, closure func() error) error {

	call := &PerformCall{Name: name}

	// lazily instantiate to make matching sub-stages easier
	if s.PerformCalls == nil {
		s.PerformCalls = []*PerformCall{}
	}
	s.PerformCalls = append(s.PerformCalls, call) // We want to record the calls in the same order as the real implementation would print them

	err := closure()

	call.Error = err
	if err != nil {
		var skipErr biui.SkipStageError
		if errors.As(err, &skipErr) {
			call.SkipError = skipErr
			err = nil
		}
	}

	return err
}

func (s *FakeStage) PerformComplex(name string, closure func(biui.Stage) error) error {
	subStage := NewFakeStage()

	// lazily instantiate to make matching simple stages easier
	if s.SubStages == nil {
		s.SubStages = []*FakeStage{}
	}
	s.SubStages = append(s.SubStages, subStage)

	err := closure(subStage)

	call := &PerformCall{Name: name, Error: err, Stage: subStage}

	if err != nil {
		var skipErr biui.SkipStageError
		if errors.As(err, &skipErr) {
			call.SkipError = skipErr
			err = nil
		}
	}

	// lazily instantiate to make matching sub-stages easier
	if s.PerformCalls == nil {
		s.PerformCalls = []*PerformCall{}
	}
	s.PerformCalls = append(s.PerformCalls, call)

	return err
}
