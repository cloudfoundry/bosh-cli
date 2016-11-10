package cmd

import boshdir "github.com/cloudfoundry/bosh-cli/director"

type AttachDiskCmd struct {
	director   boshdir.Director
	deployment boshdir.Deployment
}

func NewAttachDiskCmd(director boshdir.Director, deployment boshdir.Deployment) AttachDiskCmd {
	return AttachDiskCmd{
		director:   director,
		deployment: deployment,
	}
}

func (c AttachDiskCmd) Run(opts AttachDiskOpts) error {
	return c.director.AttachDisk(c.deployment, opts.Args.Slug, opts.Args.DiskId)
}
