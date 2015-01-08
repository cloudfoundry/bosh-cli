package pkg

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcpirel "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type ReleaseCompiler interface {
	Compile(release bmrel.Release, manifest bminstallmanifest.Manifest) error
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

func (c releaseCompiler) Compile(release bmrel.Release, manifest bminstallmanifest.Manifest) error {
	c.logger.Info(c.logTag, "Compiling CPI release '%s'", release.Name())
	c.logger.Debug(c.logTag, fmt.Sprintf("Compiling CPI release '%s': %#v", release.Name(), release))

	//TODO: should only be compiling the packages required by the cpi job
	err := c.packagesCompiler.Compile(release)
	if err != nil {
		return bosherr.WrapError(err, "Compiling release packages")
	}

	manifestProperties, err := manifest.Properties()
	if err != nil {
		return bosherr.WrapError(err, "Getting installation manifest properties")
	}

	cpiJob, found := release.FindJobByName(bmcpirel.ReleaseJobName)
	if !found {
		return bosherr.WrapErrorf(err, "Job '%s' not found in release '%s'", bmcpirel.ReleaseJobName, release.Name())
	}

	err = c.templatesCompiler.Compile([]bmrel.Job{cpiJob}, manifest.Name, manifestProperties)
	if err != nil {
		return bosherr.WrapError(err, "Compiling job templates")
	}
	return nil
}
