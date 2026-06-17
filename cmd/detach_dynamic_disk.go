package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DetachDynamicDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDetachDynamicDiskCmd(ui boshui.UI, director boshdir.Director) DetachDynamicDiskCmd {
	return DetachDynamicDiskCmd{ui: ui, director: director}
}

func (c DetachDynamicDiskCmd) Run(opts DetachDynamicDiskOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.DetachDynamicDisk(opts.Args.DiskName)
}
