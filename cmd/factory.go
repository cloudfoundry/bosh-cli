package cmd

import (
	"errors"
	"os"
	"os/user"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands map[string]Cmd
	logger   boshlog.Logger
}

func NewFactory(logger boshlog.Logger) Factory {
	usr, _ := user.Current()
	ui := bmui.NewDefaultUI(os.Stdout, os.Stderr)
	filesystem := boshsys.NewOsFileSystem(logger)

	return &factory{
		commands: map[string]Cmd{
			"deployment": NewDeploymentCmd(ui, usr.HomeDir, filesystem),
		},
	}
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name], nil
}
