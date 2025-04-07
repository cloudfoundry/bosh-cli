package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type AttachDiskCmd struct {
	deployment boshdir.Deployment
}

func NewAttachDiskCmd(deployment boshdir.Deployment) AttachDiskCmd {
	return AttachDiskCmd{
		deployment: deployment,
	}
}

func (c AttachDiskCmd) Run(opts AttachDiskOpts) error {
	return c.deployment.AttachDisk(opts.Args.Slug, opts.Args.DiskCID, opts.DiskProperties)
}
