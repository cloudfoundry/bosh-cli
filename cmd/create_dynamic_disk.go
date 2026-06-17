package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type CreateDynamicDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCreateDynamicDiskCmd(ui boshui.UI, director boshdir.Director) CreateDynamicDiskCmd {
	return CreateDynamicDiskCmd{ui: ui, director: director}
}

func (c CreateDynamicDiskCmd) Run(opts CreateDynamicDiskOpts) error {
	diskName := opts.Args.DiskName

	diskCID, err := c.director.CreateDynamicDisk(diskName, opts.DiskPool, opts.Size, nil)
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Dynamic disk '%s' created (CID: %s)", diskName, diskCID)
	return nil
}
