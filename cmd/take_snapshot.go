package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type TakeSnapshotCmd struct {
	deployment boshdir.Deployment
}

func NewTakeSnapshotCmd(deployment boshdir.Deployment) TakeSnapshotCmd {
	return TakeSnapshotCmd{deployment: deployment}
}

func (c TakeSnapshotCmd) Run(opts TakeSnapshotOpts) error {
	if opts.Args.Slug.IsProvided() {
		return c.deployment.TakeSnapshot(opts.Args.Slug)
	}

	return c.deployment.TakeSnapshots()
}
