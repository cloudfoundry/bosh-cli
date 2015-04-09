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
	if len(args) == 0 {
		return bosherr.Error("Invalid usage: No command specified")
	}

	commandName := args[0]
	cmd, err := r.factory.CreateCommand(commandName)
	if err != nil {
		return bosherr.WrapErrorf(err, "Command '%s' unknown", commandName)
	}

	err = cmd.Run(stage, args[1:])
	if err != nil {
		return bosherr.WrapErrorf(err, "Command '%s' failed", commandName)
	}

	return nil
}
