package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeleteVmCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewDeleteVmCmd(ui boshui.UI, deployment boshdir.Deployment) DeleteVmCmd {
	return DeleteVmCmd{ui: ui, deployment: deployment}
}

func (c DeleteVmCmd) Run(opts DeleteVmOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.DeleteVm(opts.Args.CID)
}
