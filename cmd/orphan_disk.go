package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type OrphanDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewOrphanDiskCmd(ui boshui.UI, director boshdir.Director) OrphanDiskCmd {
	return OrphanDiskCmd{ui: ui, director: director}
}

func (c OrphanDiskCmd) Run(opts OrphanDiskOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	disk, err := c.director.FindOrphanedDisk(opts.Args.CID)
	if err != nil {
		return err
	}

	return disk.Orphan()
}
