package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Runner struct {
	factory Factory
}

func NewRunner(factory Factory) *Runner {
	return &Runner{
		factory: factory,
	}
}

func (r *Runner) Run(stage biui.Stage, args ...string) error {
	var commandName string
	if len(args) == 0 {
		commandName = "help"
	} else {
		commandName = args[0]
		args = args[1:]
	}

	cmd, err := r.factory.CreateCommand(commandName)
	if err != nil {
		return bosherr.WrapErrorf(err, "Command '%s' unknown", commandName)
	}

	err = cmd.Run(stage, args)
	if err != nil {
		return bosherr.WrapErrorf(err, "Command '%s' failed", commandName)
	}

	return nil
}
