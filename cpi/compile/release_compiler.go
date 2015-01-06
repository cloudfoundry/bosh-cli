package compile

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type ReleaseCompiler interface {
	Compile(release bmrel.Release, deployment bminstallmanifest.Manifest) error
}

type releaseCompiler struct {
	packagesCompiler  ReleasePackagesCompiler
	templatesCompiler bmtemcomp.TemplatesCompiler
	logger            boshlog.Logger
	logTag            string
}

func NewReleaseCompiler(
	packagesCompiler ReleasePackagesCompiler,
	templatesCompiler bmtemcomp.TemplatesCompiler,
	logger boshlog.Logger,
) ReleaseCompiler {
	return &releaseCompiler{
		packagesCompiler:  packagesCompiler,
		templatesCompiler: templatesCompiler,
		logger:            logger,
		logTag:            "releaseCompiler",
	}
}

func (c releaseCompiler) Compile(release bmrel.Release, deployment bminstallmanifest.Manifest) error {
	c.logger.Info(c.logTag, "Compiling CPI release '%s'", release.Name())
	c.logger.Debug(c.logTag, fmt.Sprintf("Compiling CPI release '%s': %#v", release.Name(), release))

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
