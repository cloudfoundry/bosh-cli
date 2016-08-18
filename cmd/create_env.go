package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type CreateEnvCmd struct {
	ui          boshui.UI
	envProvider func(string, boshtpl.Variables) DeploymentPreparer
}

func NewCreateEnvCmd(ui boshui.UI, envProvider func(string, boshtpl.Variables) DeploymentPreparer) *CreateEnvCmd {
	return &CreateEnvCmd{ui: ui, envProvider: envProvider}
}

func (c *CreateEnvCmd) Run(stage boshui.Stage, opts CreateEnvOpts) error {
	c.ui.BeginLinef("Deployment manifest: '%s'\n", opts.Args.Manifest.Path)

	depPreparer := c.envProvider(opts.Args.Manifest.Path, opts.VarFlags.AsVariables())

	return depPreparer.PrepareDeployment(stage)
}
