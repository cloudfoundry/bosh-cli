package cmd

import (
	"errors"
)

type deploymentCmd struct {
	args []string
}

func NewDeploymentCmd() *deploymentCmd {
	return &deploymentCmd{}
}

func (f *deploymentCmd) Run(args []string) error {
	f.args = args

	return errors.New("Implement me!")
}
