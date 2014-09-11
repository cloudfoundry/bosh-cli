package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"

	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

const (
	logTag = "depoyCmd"
)

type deployCmd struct {
	ui              bmui.UI
	config          bmconfig.Config
	fs              boshsys.FileSystem
	extractor       boshcmd.Compressor
	validator       bmrelvalidation.ReleaseValidator
	releaseCompiler bmcomp.ReleaseCompiler
	repo            bmstemcell.Repo
	logger          boshlog.Logger
}

func NewDeployCmd(
	ui bmui.UI,
	config bmconfig.Config,
	fs boshsys.FileSystem,
	extractor boshcmd.Compressor,
	validator bmrelvalidation.ReleaseValidator,
	releaseCompiler bmcomp.ReleaseCompiler,
	repo bmstemcell.Repo,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:              ui,
		config:          config,
		fs:              fs,
		extractor:       extractor,
		validator:       validator,
		releaseCompiler: releaseCompiler,
		repo:            repo,
		logger:          logger,
	}
}

func (c *deployCmd) Run(args []string) error {
	releaseTarballPath, stemcellTarballPath, err := c.validateDeployInputs(args)
	if err != nil {
		return err
	}

	//TODO: extract deployment parsing from ReleaseCompiler.Compile
	c.parseCPIDeploymentManifest()

	cloud, err := c.deployCPIDeployment(c.config.Deployment, releaseTarballPath)
	if err != nil {
		return err
	}

	stemcell, err := c.uploadStemcell(cloud, stemcellTarballPath)
	if err != nil {
		return err
	}

	microboshDeployment, err := c.parseMicroboshManifest()
	if err != nil {
		return err
	}

	err = c.deployMicrobosh(cloud, microboshDeployment, stemcell)
	if err != nil {
		return err
	}

	// register the stemcell
	return nil
}

type CPIDeployment struct{}
type Cloud struct{}
type Deployment struct{}

//func (c *deployCmd) Run(args []string) error {
//  releaseTarballPath, stemcellTarballPath := c.validateDeployInputs(args)
//
//  cpiDeployment := c.parseCPIDeploymentManifest()
//  cloud := c.deployLocalDeployment(cpiDeployment, releaseTarballPath)
//
//  stemcell := c.uploadStemcell(cloud, stemcellTarballPath)
//  microboshDeployment := c.parseMicroboshManifest()
//  c.deployMicrobosh(cloud, microboshDeployment, stemcell)
//}

// validateDeployInputs validates the presence of inputs (stemcell tarball, cpi release tarball)
func (c *deployCmd) validateDeployInputs(args []string) (string, string, error) {

	if len(args) == 0 {
		c.ui.Error("No CPI release provided")
		c.logger.Debug(logTag, "No CPI release provided")
		return "", "", errors.New("No CPI release provided")
	}

	releaseTarballPath := args[0]
	c.logger.Info(logTag, fmt.Sprintf("Validating deployment `%s'", releaseTarballPath))

	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(releaseTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' does not exist", releaseTarballPath))
		return "", "", bosherr.WrapError(err, "Checking CPI release `%s' existence", releaseTarballPath)
	}

	// validate current state: 'microbosh' deployment set
	if len(c.config.Deployment) == 0 {
		c.ui.Error("No deployment set")
		return "", "", bosherr.New("No deployment set")
	}

	c.logger.Info(logTag, fmt.Sprintf("Checking for deployment `%s'", c.config.Deployment))
	err = fileValidator.Exists(c.config.Deployment)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path `%s' does not exist", c.config.Deployment))
		return "", "", bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	stemcellTarballPath := args[1]
	//TODO Validate existence of stemcellTarballPath

	return releaseTarballPath, stemcellTarballPath, nil

}

func (c *deployCmd) parseCPIDeploymentManifest() (deployment CPIDeployment) {
	//c.config.Deployment
	return CPIDeployment{}
}

func (c *deployCmd) deployCPIDeployment(deploymentPath string, releaseTarballPath string) (Cloud, error) {
	// unpack cpi release source
	c.logger.Info(logTag, "Extracting CPI release")
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
	c.logger.Info(logTag, "Extracted CPI release to `%s'", extractedReleasePath)

	// validate cpi release source
	c.logger.Info(logTag, "Validating CPI release")
	err = c.validator.Validate(release)
	if err != nil {
		return Cloud{}, bosherr.WrapError(err, "Validating CPI release")
	}

	c.logger.Info(logTag, fmt.Sprintf("Compiling CPI release `%s'", release.Name))
	c.logger.Debug(logTag, fmt.Sprintf("Compiling CPI release: %#v", release))

	// compile cpi release packages & store in compiled package repo
	err = c.releaseCompiler.Compile(release, deploymentPath)
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

func (c *deployCmd) uploadStemcell(_ Cloud, stemcellTarballPath string) (bmstemcell.Stemcell, error) {
	//   unpack stemcell tarball & cloud.create_stemcell(image_path)
	stemcell, extractedPath, err := c.repo.Save(stemcellTarballPath)
	if err != nil {
		c.ui.Error("Could not read stemcell")
		return bmstemcell.Stemcell{}, bosherr.WrapError(err, "Saving stemcell")
	}

	c.fs.RemoveAll(extractedPath)
	return stemcell, nil
}

func (c *deployCmd) parseMicroboshManifest() (Deployment, error) {
	//c.config.Deployment
	return Deployment{}, nil
}

func (c *deployCmd) deployMicrobosh(cpi Cloud, deployment Deployment, stemcell bmstemcell.Stemcell) error {
	// create (or discover & update) remote deployment 'cells'
	//   cloud.create_vm & store agent_id
	//   wait for agent to bootstrap
	//   tell remote agent to apply state
	//   poll agent task get_state until finished
	return nil
}
