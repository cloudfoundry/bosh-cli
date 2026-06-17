package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type ProvideDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewProvideDiskCmd(ui boshui.UI, director boshdir.Director) ProvideDiskCmd {
	return ProvideDiskCmd{ui: ui, director: director}
}

func (c ProvideDiskCmd) Run(opts ProvideDiskOpts) error {
	instanceID := opts.Args.InstanceID.String()
	diskName := opts.Args.DiskName

	diskCID, err := c.director.ProvideDynamicDisk(instanceID, diskName, opts.DiskPool, opts.Size, nil)
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Disk '%s' provided (CID: %s)", diskName, diskCID)
	return nil
}
