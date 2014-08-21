package cmd

import (
	"errors"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

type Runner struct {
	factory Factory
	args    []string
}

func NewRunner(factory Factory) *Runner {
	return &Runner{factory: factory}
}

func (runner *Runner) Run(args []string) error {
	runner.args = args

	if runner.args == nil {
		return errors.New("Invalid args, cannot be nil")
	}

	if len(runner.args) == 0 {
		return errors.New("Invalid args, cannot be empty")
	}

	commandName := args[0]
	cmd, err := runner.factory.CreateCommand(commandName)
	if err != nil {
		return bosherr.WrapError(err, "Failed creating command with name: %s", commandName)
	}

	return cmd.Run(args[1:])
}
