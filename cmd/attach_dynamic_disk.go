package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type AttachDynamicDiskCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewAttachDynamicDiskCmd(ui boshui.UI, director boshdir.Director) AttachDynamicDiskCmd {
	return AttachDynamicDiskCmd{ui: ui, director: director}
}

func (c AttachDynamicDiskCmd) Run(opts AttachDynamicDiskOpts) error {
	instanceID := opts.Args.InstanceID.String()
	diskName := opts.Args.DiskName

	return c.director.AttachDynamicDisk(diskName, instanceID)
}
