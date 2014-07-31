package cmd

import (
	"errors"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands map[string]Cmd
}

func NewFactory(
	filesystem boshsys.FileSystem,
	ui bmui.UI,
	boshMicroFilePath string,
) Factory {
	return &factory{
		commands: map[string]Cmd{
			"deployment": NewDeploymentCmd(ui, boshMicroFilePath, filesystem),
		},
	}
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name], nil
}
