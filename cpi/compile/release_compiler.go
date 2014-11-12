package compile

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type ReleaseCompiler interface {
	Compile(release bmrel.Release, deployment bmdepl.Deployment) error
}

type releaseCompiler struct {
	packagesCompiler  ReleasePackagesCompiler
	templatesCompiler bmtemcomp.TemplatesCompiler
}

func NewReleaseCompiler(
	packagesCompiler ReleasePackagesCompiler,
	templatesCompiler bmtemcomp.TemplatesCompiler,
) ReleaseCompiler {
	return &releaseCompiler{
		packagesCompiler:  packagesCompiler,
		templatesCompiler: templatesCompiler,
	}
}

func (c releaseCompiler) Compile(release bmrel.Release, deployment bmdepl.Deployment) error {
	err := c.packagesCompiler.Compile(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release packages")
	}

	err = c.templatesCompiler.Compile(release.Jobs(), deployment)
	if err != nil {
		return bosherr.WrapError(err, "Compiling job templates")
	}
	return nil
}
