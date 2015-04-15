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
	args = processHelp(args)
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

func processHelp(args []string) []string {
	if len(args) == 0 {
		return []string{"help"}
	}

	for i, arg := range args {
		if arg == "help" || arg == "-h" || arg == "-help" || arg == "--help" {
			return append(append([]string{"help"}, args[:i]...), args[i+1:]...)
		}
	}

	return args
}
