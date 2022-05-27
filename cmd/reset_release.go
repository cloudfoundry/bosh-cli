package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v6/cmd/opts"
	boshreldir "github.com/cloudfoundry/bosh-cli/v6/releasedir"
)

type ResetReleaseCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewResetReleaseCmd(releaseDir boshreldir.ReleaseDir) ResetReleaseCmd {
	return ResetReleaseCmd{releaseDir: releaseDir}
}

func (c ResetReleaseCmd) Run(opts ResetReleaseOpts) error {
	return c.releaseDir.Reset()
}
