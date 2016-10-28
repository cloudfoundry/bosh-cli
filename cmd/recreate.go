package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type RecreateCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewRecreateCmd(ui boshui.UI, deployment boshdir.Deployment) RecreateCmd {
	return RecreateCmd{ui: ui, deployment: deployment}
}

func (c RecreateCmd) Run(opts RecreateOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.Recreate(opts.Args.Slug, opts.SkipDrain, opts.Force, opts.Canaries, opts.MaxInFlight)
}
