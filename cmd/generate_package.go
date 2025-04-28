package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

type GeneratePackageCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewGeneratePackageCmd(releaseDir boshreldir.ReleaseDir) GeneratePackageCmd {
	return GeneratePackageCmd{releaseDir: releaseDir}
}

func (c GeneratePackageCmd) Run(opts GeneratePackageOpts) error {
	return c.releaseDir.GeneratePackage(opts.Args.Name)
}
