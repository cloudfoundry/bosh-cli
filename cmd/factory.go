package cmd

import (
	"errors"
	"os"
	"os/user"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands map[string]Cmd
}

func NewFactory() Factory {
	usr, _ := user.Current()
	ui := bmui.NewDefaultUI(os.Stdout, os.Stderr)

	return &factory{

		commands: map[string]Cmd{
			"deployment": NewDeploymentCmd(ui, usr.HomeDir),
		},
	}
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name], nil
}
