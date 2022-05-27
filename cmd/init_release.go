package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

type InitReleaseCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewInitReleaseCmd(releaseDir boshreldir.ReleaseDir) InitReleaseCmd {
	return InitReleaseCmd{releaseDir: releaseDir}
}

func (c InitReleaseCmd) Run(opts InitReleaseOpts) error {
	return c.releaseDir.Init(opts.Git)
}
