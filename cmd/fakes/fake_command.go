package fakes

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	fakes "github.com/cloudfoundry/bosh-agent/system/fakes"
	cmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
)

type FakeCommand struct {
	Name string
	Args []string

	PresetError error

	FakeFileSystem *fakes.FakeFileSystem
}

func (f *FakeCommand) CreateFakeCommand(name string) cmd.Cmd {
	return &FakeCommand{Name: name,
		Args:           []string{},
		PresetError:    nil,
		FakeFileSystem: &fakes.FakeFileSystem{},
	}
}

func (f *FakeCommand) FileSystem() boshsys.FileSystem {
	return f.FakeFileSystem
}

func (f *FakeCommand) Run(args []string) error {
	f.Args = args
	return f.PresetError
}

func (f *FakeCommand) GetArgs() []string {
	return f.Args
}
