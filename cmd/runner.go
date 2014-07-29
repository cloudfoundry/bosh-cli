package cmd

import (
	"errors"
	"fmt"
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
		return errors.New(fmt.Sprintf("Failed creating command with name: ", commandName))
	}

	return cmd.Run(args[1:])
}
