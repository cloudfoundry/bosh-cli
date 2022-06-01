package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DeleteDeploymentCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewDeleteDeploymentCmd(ui boshui.UI, deployment boshdir.Deployment) DeleteDeploymentCmd {
	return DeleteDeploymentCmd{ui: ui, deployment: deployment}
}

func (c DeleteDeploymentCmd) Run(opts DeleteDeploymentOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.Delete(opts.Force)
}
