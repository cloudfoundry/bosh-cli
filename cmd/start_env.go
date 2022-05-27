package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v6/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/v6/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/v6/ui"
	"github.com/cppforlife/go-patch/patch"
)

type StartEnvCmd struct {
	ui          boshui.UI
	envProvider func(string, string, boshtpl.Variables, patch.Op) DeploymentStateManager
}

func NewStartEnvCmd(ui boshui.UI, envProvider func(string, string, boshtpl.Variables, patch.Op) DeploymentStateManager) *StartEnvCmd {
	return &StartEnvCmd{ui: ui, envProvider: envProvider}
}

func (c *StartEnvCmd) Run(stage boshui.Stage, opts StartEnvOpts) error {
	c.ui.BeginLinef("Deployment manifest: '%s'\n", opts.Args.Manifest.Path)

	depStateManager := c.envProvider(
		opts.Args.Manifest.Path, opts.StatePath, opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp())

	return depStateManager.StartDeployment(stage)
}
