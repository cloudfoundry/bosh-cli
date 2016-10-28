package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type StopCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewStopCmd(ui boshui.UI, deployment boshdir.Deployment) StopCmd {
	return StopCmd{ui: ui, deployment: deployment}
}

func (c StopCmd) Run(opts StopOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.Stop(opts.Args.Slug, opts.Hard, opts.SkipDrain, opts.Force, opts.Canaries, opts.MaxInFlight)
}
