package fakes

import (
	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

type FakeCommand struct {
	Name        string
	Args        []string
	PresetError error
}

func (f *FakeCommand) CreateFakeCommand(name string) cmd.Cmd {
	return &FakeCommand{Name: name,
		Args:        []string{},
		PresetError: nil,
	}
}

func (f *FakeCommand) Run(args []string) error {
	f.Args = args
	return f.PresetError
}

func (f *FakeCommand) GetArgs() []string {
	return f.Args
}
