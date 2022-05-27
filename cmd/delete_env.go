package cmd

import (
	"github.com/cppforlife/go-patch/patch"

	. "github.com/cloudfoundry/bosh-cli/v6/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/v6/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/v6/ui"
)

type DeleteEnvCmd struct {
	ui          boshui.UI
	envProvider func(string, string, boshtpl.Variables, patch.Op) DeploymentDeleter
}

func NewDeleteEnvCmd(ui boshui.UI, envProvider func(string, string, boshtpl.Variables, patch.Op) DeploymentDeleter) *DeleteEnvCmd {
	return &DeleteEnvCmd{ui: ui, envProvider: envProvider}
}

func (c *DeleteEnvCmd) Run(stage boshui.Stage, opts DeleteEnvOpts) error {
	c.ui.BeginLinef("Deployment manifest: '%s'\n", opts.Args.Manifest.Path)

	depDeleter := c.envProvider(
		opts.Args.Manifest.Path, opts.StatePath, opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp())

	return depDeleter.DeleteDeployment(opts.SkipDrain, stage)
}
