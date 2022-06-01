package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

type GenerateJobCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewGenerateJobCmd(releaseDir boshreldir.ReleaseDir) GenerateJobCmd {
	return GenerateJobCmd{releaseDir: releaseDir}
}

func (c GenerateJobCmd) Run(opts GenerateJobOpts) error {
	return c.releaseDir.GenerateJob(opts.Args.Name)
}
