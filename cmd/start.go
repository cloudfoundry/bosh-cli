package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type StartCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewStartCmd(ui boshui.UI, deployment boshdir.Deployment) StartCmd {
	return StartCmd{ui: ui, deployment: deployment}
}

func (c StartCmd) Run(opts StartOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	changeJobStateOpts := boshdir.ChangeJobStateOpts{
		Canaries:    opts.Canaries,
		MaxInFlight: opts.MaxInFlight,
	}
	return c.deployment.Start(opts.Args.Slug, changeJobStateOpts)
}
