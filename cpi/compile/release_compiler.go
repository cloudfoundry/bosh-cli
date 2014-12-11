package compile

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type ReleaseCompiler interface {
	Compile(release bmrel.Release, deployment bmdepl.CPIDeploymentManifest) error
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

func (c releaseCompiler) Compile(release bmrel.Release, deployment bmdepl.CPIDeploymentManifest) error {
	err := c.packagesCompiler.Compile(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release packages")
	}

	deploymentProperties, err := deployment.Properties()
	if err != nil {
		return bosherr.WrapError(err, "Getting deployment properties")
	}

	err = c.templatesCompiler.Compile(release.Jobs(), deployment.Name, deploymentProperties)
	if err != nil {
		return bosherr.WrapError(err, "Compiling job templates")
	}
	return nil
}
