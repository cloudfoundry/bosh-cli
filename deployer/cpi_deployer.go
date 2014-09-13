package deployer

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/install"
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
	jobInstaller    bminstall.JobInstaller
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
	logger boshlog.Logger,
) CpiDeployer {
	return &cpiDeployer{
		ui:              ui,
		fs:              fs,
		extractor:       extractor,
		validator:       validator,
		releaseCompiler: releaseCompiler,
		jobInstaller:    jobInstaller,
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

	//TODO: Inject reader with fs & extractor pre-configured
	releaseReader := bmrel.NewReader(releaseTarballPath, extractedReleasePath, c.fs, c.extractor)
	//TODO: refactor Read to take releaseTarballPath & extractedReleasePath
	release, err := releaseReader.Read()
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release at `%s' is not a BOSH release", releaseTarballPath))
		return Cloud{}, bosherr.WrapError(err, fmt.Sprintf("Reading CPI release from `%s'", releaseTarballPath))
	}

	release.TarballPath = releaseTarballPath
	c.logger.Info(c.logTag, "Extracted CPI release `%s' to `%s'", release.Name, extractedReleasePath)

	// validate cpi release source
	c.logger.Info(c.logTag, "Validating CPI release `%s'", release.Name)
	err = c.validator.Validate(release)
	if err != nil {
		return Cloud{}, bosherr.WrapError(err, "Validating CPI release `%s'", release.Name)
	}

	//TODO: inject release name into deployment job templates
	//	for _, deploymentJob := range deployment.Jobs() {
	//		for _, jobRef := range deploymentJob.Templates() {
	//			//TODO: jobRef.SetRelease(release.Name())
	//		}
	//	}

	// compile packages & render job templates
	c.logger.Info(c.logTag, fmt.Sprintf("Compiling CPI release `%s'", release.Name))
	c.logger.Debug(c.logTag, fmt.Sprintf("Compiling CPI release `%s': %#v", release.Name, release))
	err = c.releaseCompiler.Compile(release, deployment)
	if err != nil {
		c.ui.Error("Could not compile CPI release")
		return Cloud{}, bosherr.WrapError(err, "Compiling CPI release")
	}

	// cpi deployment should only have one job (because it's a local deployment)
	jobs := deployment.Jobs
	if len(jobs) != 1 {
		c.ui.Error("Invalid CPI deployment: exactly one job required")
		return Cloud{}, bosherr.New("Invalid CPI deployment: exactly one job required, %d jobs found", len(jobs))
	}
	cpiJob := jobs[0]

	// local deployment job should only ever have 1 instance
	instances := cpiJob.Instances
	if instances != 1 {
		c.ui.Error("Invalid CPI deployment: exactly one job instance required")
		return Cloud{}, bosherr.New(
			"Invalid CPI deployment: exactly one instance required, found %d instances in job `%s'",
			instances,
			cpiJob.Name,
		)
	}

	err = c.installJob(cpiJob, release)
	if err != nil {
		c.ui.Error("Could not compile CPI release")
		return Cloud{}, bosherr.WrapError(err, "Compiling CPI release")
	}

	//TODO: delete cpi source?
	return Cloud{}, nil
}

// installJob installs the deployment job's rendered job templates & required compiled packages
// all job templates must be in the specified release
func (c *cpiDeployer) installJob(deploymentJob bmdepl.Job, release bmrel.Release) error {
	//	deploymentJobName := deploymentJob.Name()

	for _, releaseJobRef := range deploymentJob.Templates {
		// Microbosh manifests do not know the name of the cpi release...
		//TODO: uncomment after release name injection is added
		//    releaseName := releaseJobRef.Release()
		//		if releaseName != release.Name {
		//			c.ui.Error(fmt.Sprintf("Could not find release `%s', specified by job `%s', expected `%s'", releaseName, deploymentJobName, release.Name))
		//			return bosherr.New("Invalid CPI deployment manifest: release `%s' not found, specified by job `%s', expected `%s'", releaseName, deploymentJobName, release.Name)
		//		}

		releaseJobName := releaseJobRef.Name
		releaseJob, found := release.FindJobByName(releaseJobName)

		if !found {
			c.ui.Error(fmt.Sprintf("Could not find CPI job `%s' in release `%s'", releaseJobName, release.Name))
			return bosherr.New("Invalid CPI deployment manifest: job `%s' not found in release `%s'", releaseJobName, release.Name)
		}

		err := c.jobInstaller.Install(releaseJob)
		if err != nil {
			c.ui.Error(fmt.Sprintf("Could not install `%s' job", releaseJobName))
			return bosherr.WrapError(err, "Installing `%s' job for CPI release", releaseJobName)
		}
	}
	return nil
}
