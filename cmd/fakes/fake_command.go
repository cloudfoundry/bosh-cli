package fakes

import (
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type FakeCommand struct {
	name        string
	Stage       bmui.Stage
	Args        []string
	PresetError error
}

func NewFakeCommand(name string) *FakeCommand {
	return &FakeCommand{
		name: name,
		Args: []string{},
	}
}

func (f *FakeCommand) Name() string {
	return f.name
}

func (f *FakeCommand) Run(stage bmui.Stage, args []string) error {
	f.Stage = stage
	f.Args = args
	return f.PresetError
}

func (f *FakeCommand) GetArgs() []string {
	return f.Args
}
