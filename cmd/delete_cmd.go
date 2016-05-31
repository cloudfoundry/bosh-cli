package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type DeleteCmd struct {
	ui          boshui.UI
	envProvider func(string, boshtpl.Variables) DeploymentDeleter
}

func NewDeleteCmd(ui boshui.UI, envProvider func(string, boshtpl.Variables) DeploymentDeleter) *DeleteCmd {
	return &DeleteCmd{ui: ui, envProvider: envProvider}
}

func (c *DeleteCmd) Run(stage boshui.Stage, opts DeleteEnvOpts) error {
	c.ui.PrintLinef("Deployment manifest: '%s'", opts.Args.Manifest.Path)

	depDeleter := c.envProvider(opts.Args.Manifest.Path, opts.VarFlags.AsVariables())

	return depDeleter.DeleteDeployment(stage)
}
