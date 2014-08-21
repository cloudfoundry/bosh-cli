package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

const (
	logTag = "depoyCmd"
)

type deployCmd struct {
	ui        bmui.UI
	config    bmconfig.Config
	fs        boshsys.FileSystem
	extractor bmtar.Extractor
	validator bmrelvalidation.ReleaseValidator
	compiler  bmcomp.ReleaseCompiler
	logger    boshlog.Logger
}

func NewDeployCmd(
	ui bmui.UI,
	config bmconfig.Config,
	fs boshsys.FileSystem,
	extractor bmtar.Extractor,
	validator bmrelvalidation.ReleaseValidator,
	compiler bmcomp.ReleaseCompiler,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:        ui,
		config:    config,
		fs:        fs,
		extractor: extractor,
		validator: validator,
		compiler:  compiler,
		logger:    logger,
	}
}

func (c *deployCmd) Run(args []string) error {
	if len(args) == 0 {
		c.ui.Error("No CPI release provided")
		c.logger.Debug(logTag, "No CPI release provided")
		return errors.New("No CPI release provided")
	}

	releaseTarballPath := args[0]
	c.logger.Info(logTag, fmt.Sprintf("Validating deployment `%s'", releaseTarballPath))
	err := c.validateDeployment(releaseTarballPath)
	if err != nil {
		return err
	}

	c.logger.Info(logTag, "Extracting release")
	extractedReleasePath, err := c.fs.TempDir("cmd-deployCmd")
	if err != nil {
		c.ui.Error("Could not create a temporary directory")
		c.logger.Error(logTag, fmt.Sprintf("Could not create a temporary directory: `%s'", err.Error()))
		return bosherr.WrapError(err, "Creating extracted release path")
	}
	defer c.fs.RemoveAll(extractedReleasePath)

	release, err := c.extractRelease(releaseTarballPath, extractedReleasePath)
	if err != nil {
		return err
	}
	c.logger.Info(logTag, "Extracted release to `%s'", extractedReleasePath)

	c.logger.Info(logTag, "Validating release")
	err = c.validator.Validate(release)
	if err != nil {
		return err
	}

	c.logger.Info(logTag, fmt.Sprintf("Compiling release `%s'", release.Name))
	err = c.compiler.Compile(release)
	if err != nil {
		c.ui.Error("Could not compile release")
		c.logger.Error(logTag, fmt.Sprintf("Could not compile release `%s': `%s'", release.Name, err.Error()))
		return bosherr.WrapError(err, "Compiling release")
	}

	return nil
}

func (c *deployCmd) validateDeployment(releaseTarballPath string) error {
	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(releaseTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' does not exist", releaseTarballPath))
		c.logger.Error(logTag, "CPI release `%s' does not exist", releaseTarballPath)
		return bosherr.WrapError(err, "Checking CPI release `%s' existence", releaseTarballPath)
	}

	if len(c.config.Deployment) == 0 {
		c.ui.Error("No deployment set")
		c.logger.Error(logTag, "No deployment set")
		return errors.New("No deployment set")
	}

	err = fileValidator.Exists(c.config.Deployment)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path `%s' does not exist", c.config.Deployment))
		c.logger.Error(logTag, fmt.Sprintf("Deployment manifest path `%s' does not exist", c.config.Deployment))
		return bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	return nil
}

func (c *deployCmd) extractRelease(releaseTarballPath, extractedReleasePath string) (bmrel.Release, error) {
	releaseReader := bmrel.NewTarReader(releaseTarballPath, extractedReleasePath, c.fs, c.extractor)

	release, err := releaseReader.Read()
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' is not a BOSH release", releaseTarballPath))
		c.logger.Error(logTag, fmt.Sprintf("CPI release `%s' is not a BOSH release", releaseTarballPath))
		return bmrel.Release{}, bosherr.WrapError(err, fmt.Sprintf("Reading CPI release from `%s'", releaseTarballPath))
	}
	release.TarballPath = releaseTarballPath

	return release, nil
}
