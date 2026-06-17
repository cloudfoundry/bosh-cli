package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DeleteDynamicDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDeleteDynamicDiskCmd(ui boshui.UI, director boshdir.Director) DeleteDynamicDiskCmd {
	return DeleteDynamicDiskCmd{ui: ui, director: director}
}

func (c DeleteDynamicDiskCmd) Run(opts DeleteDynamicDiskOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.DeleteDynamicDisk(opts.Args.DiskName)
}
