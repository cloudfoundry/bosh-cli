package cmd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	boshdavcli "github.com/cloudfoundry/bosh-agent/davcli/client"
	boshdavcliconf "github.com/cloudfoundry/bosh-agent/davcli/config"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpideploy "github.com/cloudfoundry/bosh-micro-cli/cpideployer"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmmicrodeploy "github.com/cloudfoundry/bosh-micro-cli/microdeployer"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/microdeployer/agentclient"
	bmapplyspec "github.com/cloudfoundry/bosh-micro-cli/microdeployer/applyspec"
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/microdeployer/blobstore"
	bminsup "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instanceupdater"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

type deployCmd struct {
	ui                     bmui.UI
	userConfig             bmconfig.UserConfig
	fs                     boshsys.FileSystem
	cpiManifestParser      bmdepl.ManifestParser
	boshManifestParser     bmdepl.ManifestParser
	cpiDeployer            bmcpideploy.CpiDeployer
	stemcellManagerFactory bmstemcell.ManagerFactory
	microDeployer          bmmicrodeploy.Deployer
	compressor             boshcmd.Compressor
	erbrenderer            bmerbrenderer.ERBRenderer
	uuidGenerator          boshuuid.Generator
	deploymentUUID         string
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployCmd(
	ui bmui.UI,
	userConfig bmconfig.UserConfig,
	fs boshsys.FileSystem,
	cpiManifestParser bmdepl.ManifestParser,
	boshManifestParser bmdepl.ManifestParser,
	cpiDeployer bmcpideploy.CpiDeployer,
	stemcellManagerFactory bmstemcell.ManagerFactory,
	microDeployer bmmicrodeploy.Deployer,
	compressor boshcmd.Compressor,
	erbrenderer bmerbrenderer.ERBRenderer,
	uuidGenerator boshuuid.Generator,
	deploymentUUID string,
	logger boshlog.Logger,
) *deployCmd {
	return &deployCmd{
		ui:                     ui,
		userConfig:             userConfig,
		fs:                     fs,
		cpiManifestParser:      cpiManifestParser,
		boshManifestParser:     boshManifestParser,
		cpiDeployer:            cpiDeployer,
		stemcellManagerFactory: stemcellManagerFactory,
		microDeployer:          microDeployer,
		compressor:             compressor,
		erbrenderer:            erbrenderer,
		uuidGenerator:          uuidGenerator,
		deploymentUUID:         deploymentUUID,
		logger:                 logger,
		logTag:                 "deployCmd",
	}
}

func (c *deployCmd) Name() string {
	return "deploy"
}

func (c *deployCmd) Run(args []string) error {
	releaseTarballPath, stemcellTarballPath, err := c.validateDeployInputs(args)
	if err != nil {
		return err
	}

	cpiDeployment, err := c.cpiManifestParser.Parse(c.userConfig.DeploymentFile)
	if err != nil {
		return bosherr.WrapError(err, "Parsing CPI deployment manifest `%s'", c.userConfig.DeploymentFile)
	}

	boshDeployment, err := c.boshManifestParser.Parse(c.userConfig.DeploymentFile)
	if err != nil {
		return bosherr.WrapError(err, "Parsing Bosh deployment manifest `%s'", c.userConfig.DeploymentFile)
	}

	cloud, err := c.cpiDeployer.Deploy(cpiDeployment, releaseTarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Deploying CPI `%s'", releaseTarballPath)
	}

	stemcellManager := c.stemcellManagerFactory.NewManager(cloud)
	stemcell, stemcellCID, err := stemcellManager.Upload(stemcellTarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Uploading stemcell from `%s'", stemcellTarballPath)
	}

	agentClient := bmagentclient.NewAgentClient(cpiDeployment.Mbus, c.deploymentUUID, 1*time.Second, c.logger)
	agentPingRetryable := bmagentclient.NewPingRetryable(agentClient)
	agentPingRetryStrategy := bmretrystrategy.NewAttemptRetryStrategy(300, 500*time.Millisecond, agentPingRetryable, c.logger)
	endpoint, username, password, err := cpiDeployment.MbusConfig()
	if err != nil {
		return bosherr.WrapError(err, "Creating blobstore config")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := http.Client{Transport: tr}

	davClient := boshdavcli.NewClient(boshdavcliconf.Config{
		Endpoint: fmt.Sprintf("%s/blobs", endpoint),
		User:     username,
		Password: password,
	}, &httpClient)

	blobstore := bmblobstore.NewBlobstore(davClient, c.fs, c.logger)
	sha1Calculator := bmapplyspec.NewSha1Calculator(c.fs)
	applySpecCreator := bminsup.NewApplySpecCreator(sha1Calculator)
	instanceUpdater := bminsup.NewInstanceUpdater(
		agentClient,
		stemcell.ApplySpec,
		boshDeployment,
		blobstore,
		c.compressor,
		c.erbrenderer,
		c.uuidGenerator,
		applySpecCreator,
		c.fs,
		c.logger,
	)

	err = c.microDeployer.Deploy(
		cloud,
		boshDeployment,
		cpiDeployment.Registry,
		cpiDeployment.SSHTunnel,
		agentPingRetryStrategy,
		stemcellCID,
		instanceUpdater,
	)
	if err != nil {
		return bosherr.WrapError(err, "Deploying Microbosh")
	}

	// register the stemcell
	return nil
}

type Deployment struct{}

// validateDeployInputs validates the presence of inputs (stemcell tarball, cpi release tarball)
func (c *deployCmd) validateDeployInputs(args []string) (string, string, error) {

	if len(args) != 2 {
		c.ui.Error("Invalid usage - deploy command requires exactly 2 arguments")
		c.ui.Sayln("Expected usage: bosh-micro deploy <cpi-release-tarball> <stemcell-tarball>")
		c.logger.Error(c.logTag, "Invalid arguments: ")
		return "", "", errors.New("Invalid usage - deploy command requires exactly 2 arguments")
	}

	releaseTarballPath := args[0]
	c.logger.Info(c.logTag, "Validating release tarball `%s'", releaseTarballPath)

	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(releaseTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release `%s' does not exist", releaseTarballPath))
		return "", "", bosherr.WrapError(err, "Checking CPI release `%s' existence", releaseTarballPath)
	}

	stemcellTarballPath := args[1]
	c.logger.Info(c.logTag, "Validating stemcell tarball `%s'", stemcellTarballPath)
	err = fileValidator.Exists(stemcellTarballPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Stemcell `%s' does not exist", stemcellTarballPath))
		return "", "", bosherr.WrapError(err, "Checking stemcell `%s' existence", stemcellTarballPath)
	}

	// validate current state: 'microbosh' deployment set
	if len(c.userConfig.DeploymentFile) == 0 {
		c.ui.Error("No deployment set")
		return "", "", bosherr.New("No deployment set")
	}

	c.logger.Info(c.logTag, "Checking for deployment `%s'", c.userConfig.DeploymentFile)
	err = fileValidator.Exists(c.userConfig.DeploymentFile)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path `%s' does not exist", c.userConfig.DeploymentFile))
		return "", "", bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	return releaseTarballPath, stemcellTarballPath, nil
}
