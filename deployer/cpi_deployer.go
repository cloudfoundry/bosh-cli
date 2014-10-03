package deployer

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/install"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type CpiDeployer interface {
	Deploy(deployment bmdepl.Deployment, releaseTarballPath string) (bmcloud.Cloud, error)
}

type cpiDeployer struct {
	ui              bmui.UI
	fs              boshsys.FileSystem
	extractor       boshcmd.Compressor
	validator       bmrelvalidation.ReleaseValidator
	releaseCompiler bmcomp.ReleaseCompiler
	jobInstaller    bminstall.JobInstaller
	cloudFactory    bmcloud.Factory
	logger          boshlog.Logger
	logTag          string
}

func NewCpiDeployer(
	ui bmui.UI,
	fs boshsys.FileSystem,
	extractor boshcmd.Compressor,
	validator bmrelvalidation.ReleaseValidator,
	releaseCompiler bmcomp.ReleaseCompiler,
	jobInstaller bminstall.JobInstaller,
	cloudFactory bmcloud.Factory,
	logger boshlog.Logger,
) CpiDeployer {
	return &cpiDeployer{
		ui:              ui,
		fs:              fs,
		extractor:       extractor,
		validator:       validator,
		releaseCompiler: releaseCompiler,
		jobInstaller:    jobInstaller,
		cloudFactory:    cloudFactory,
		logger:          logger,
		logTag:          "cpiDeployer",
	}
}

func (c *cpiDeployer) Deploy(deployment bmdepl.Deployment, releaseTarballPath string) (bmcloud.Cloud, error) {
	c.logger.Info(c.logTag, "Extracting CPI release")
	extractedReleasePath, err := c.fs.TempDir("cmd-deployCmd")
	if err != nil {
		c.ui.Error("Could not create a temporary directory")
		return nil, bosherr.WrapError(err, "Creating temp directory")
	}
	defer c.fs.RemoveAll(extractedReleasePath)

	releaseReader := bmrel.NewReader(releaseTarballPath, extractedReleasePath, c.fs, c.extractor)
	release, err := releaseReader.Read()
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release at `%s' is not a BOSH release", releaseTarballPath))
		return nil, bosherr.WrapError(err, fmt.Sprintf("Reading CPI release from `%s'", releaseTarballPath))
	}

	release.TarballPath = releaseTarballPath
	c.logger.Info(c.logTag, "Extracted CPI release `%s' to `%s'", release.Name, extractedReleasePath)

	c.logger.Info(c.logTag, "Validating CPI release `%s'", release.Name)
	err = c.validator.Validate(release)
	if err != nil {
		return nil, bosherr.WrapError(err, "Validating CPI release `%s'", release.Name)
	}

	c.logger.Info(c.logTag, fmt.Sprintf("Compiling CPI release `%s'", release.Name))
	c.logger.Debug(c.logTag, fmt.Sprintf("Compiling CPI release `%s': %#v", release.Name, release))
	err = c.releaseCompiler.Compile(release, deployment)
	if err != nil {
		c.ui.Error("Could not compile CPI release")
		return nil, bosherr.WrapError(err, "Compiling CPI release")
	}

	jobs := deployment.Jobs
	if len(jobs) != 1 {
		c.ui.Error("Invalid CPI deployment: exactly one job required")
		return nil, bosherr.New("Invalid CPI deployment: exactly one job required, %d jobs found", len(jobs))
	}
	cpiJob := jobs[0]

	instances := cpiJob.Instances
	if instances != 1 {
		c.ui.Error("Invalid CPI deployment: exactly one job instance required")
		return nil, bosherr.New(
			"Invalid CPI deployment: exactly one instance required, found %d instances in job `%s'",
			instances,
			cpiJob.Name,
		)
	}

	installedJobs, err := c.installJob(cpiJob, release)
	if err != nil {
		c.ui.Error("Could not install CPI deployment job")
		return nil, bosherr.WrapError(err, "Installing CPI deployment job")
	}

	cloud, err := c.cloudFactory.NewCloud(installedJobs)
	if err != nil {
		c.ui.Error("Invalid CPI deployment")
		return nil, bosherr.WrapError(err, "Validating CPI deployment job installation")
	}

	return cloud, nil
}

func (c *cpiDeployer) installJob(deploymentJob bmdepl.Job, release bmrel.Release) ([]bminstall.InstalledJob, error) {
	installedJobs := make([]bminstall.InstalledJob, 0, len(deploymentJob.Templates))
	for _, releaseJobRef := range deploymentJob.Templates {
		releaseJobName := releaseJobRef.Name
		releaseJob, found := release.FindJobByName(releaseJobName)

		if !found {
			c.ui.Error(fmt.Sprintf("Could not find CPI job `%s' in release `%s'", releaseJobName, release.Name))
			return installedJobs, bosherr.New("Invalid CPI deployment manifest: job `%s' not found in release `%s'", releaseJobName, release.Name)
		}

		installedJob, err := c.jobInstaller.Install(releaseJob)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Could not install `%s' job", releaseJobName))
			return installedJobs, bosherr.WrapError(err, "Installing `%s' job for CPI release", releaseJobName)
		}
		installedJobs = append(installedJobs, installedJob)
	}
	return installedJobs, nil
}
