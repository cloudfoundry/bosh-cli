package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DeleteSnapshotCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewDeleteSnapshotCmd(ui boshui.UI, deployment boshdir.Deployment) DeleteSnapshotCmd {
	return DeleteSnapshotCmd{ui: ui, deployment: deployment}
}

func (c DeleteSnapshotCmd) Run(opts DeleteSnapshotOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.DeleteSnapshot(opts.Args.CID)
}
