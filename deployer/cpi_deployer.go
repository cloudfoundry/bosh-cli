package deployer

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type CpiDeployer interface {
	Deploy(deployment bmdepl.Deployment, releaseTarballPath string) (Cloud, error)
}

type cpiDeployer struct {
	ui              bmui.UI
	fs              boshsys.FileSystem
	extractor       boshcmd.Compressor
	validator       bmrelvalidation.ReleaseValidator
	releaseCompiler bmcomp.ReleaseCompiler
	logger          boshlog.Logger
	logTag          string
}

func NewCpiDeployer(
	ui bmui.UI,
	fs boshsys.FileSystem,
	extractor boshcmd.Compressor,
	validator bmrelvalidation.ReleaseValidator,
	releaseCompiler bmcomp.ReleaseCompiler,
	logger boshlog.Logger,
) CpiDeployer {
	return &cpiDeployer{
		ui:              ui,
		fs:              fs,
		extractor:       extractor,
		validator:       validator,
		releaseCompiler: releaseCompiler,
		logger:          logger,
		logTag:          "cpiDeployer",
	}
}

func (c *cpiDeployer) Deploy(deployment bmdepl.Deployment, releaseTarballPath string) (Cloud, error) {
	// unpack cpi release source
	c.logger.Info(c.logTag, "Extracting CPI release")
	extractedReleasePath, err := c.fs.TempDir("cmd-deployCmd")
	if err != nil {
		c.ui.Error("Could not create a temporary directory")
		return Cloud{}, bosherr.WrapError(err, "Creating temp directory")
	}
	defer c.fs.RemoveAll(extractedReleasePath)

	releaseReader := bmrel.NewReader(releaseTarballPath, extractedReleasePath, c.fs, c.extractor)
	release, err := releaseReader.Read()
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' is not a BOSH release", releaseTarballPath))
		return Cloud{}, bosherr.WrapError(err, fmt.Sprintf("Reading CPI release from `%s'", releaseTarballPath))
	}
	release.TarballPath = releaseTarballPath
	c.logger.Info(c.logTag, "Extracted CPI release to `%s'", extractedReleasePath)

	// validate cpi release source
	c.logger.Info(c.logTag, "Validating CPI release")
	err = c.validator.Validate(release)
	if err != nil {
		return Cloud{}, bosherr.WrapError(err, "Validating CPI release")
	}

	c.logger.Info(c.logTag, fmt.Sprintf("Compiling CPI release `%s'", release.Name))
	c.logger.Debug(c.logTag, fmt.Sprintf("Compiling CPI release: %#v", release))

	// compile cpi release packages & store in compiled package repo
	err = c.releaseCompiler.Compile(release, deployment)
	if err != nil {
		c.ui.Error("Could not compile CPI release")
		return Cloud{}, bosherr.WrapError(err, "Compiling CPI release")
	}

	// create local deployment 'cells'
	//   tell 'local agent' to apply state
	//   install compiled packages
	//   install compiled job templates
	// delete cpi source
	return Cloud{}, nil
}
