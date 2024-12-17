package cmd

import (
	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/go-patch/patch"

	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bidepl "github.com/cloudfoundry/bosh-cli/v7/deployment"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DeploymentStateManager interface {
	StopDeployment(skipDrain bool, stage biui.Stage) (err error)
	StartDeployment(stage biui.Stage) (err error)
}

func NewDeploymentStateManager(
	ui biui.UI,
	logTag string,
	logger boshlog.Logger,
	deploymentStateService biconfig.DeploymentStateService,
	agentClientFactory bihttpagent.AgentClientFactory,
	deploymentManagerFactory bidepl.ManagerFactory,
	deploymentManifestPath string,
	deploymentVars boshtpl.Variables,
	deploymentOp patch.Op,
	releaseSetAndInstallationManifestParser ReleaseSetAndInstallationManifestParser,
	deploymentManifestParser DeploymentManifestParser,
) DeploymentStateManager {
	return &deploymentStateManager{
		ui:                                      ui,
		logTag:                                  logTag,
		logger:                                  logger,
		deploymentStateService:                  deploymentStateService,
		agentClientFactory:                      agentClientFactory,
		deploymentManagerFactory:                deploymentManagerFactory,
		deploymentManifestPath:                  deploymentManifestPath,
		deploymentVars:                          deploymentVars,
		deploymentOp:                            deploymentOp,
		releaseSetAndInstallationManifestParser: releaseSetAndInstallationManifestParser,
		deploymentManifestParser:                deploymentManifestParser,
	}
}

type deploymentStateManager struct {
	ui                                      biui.UI
	logTag                                  string
	logger                                  boshlog.Logger
	deploymentStateService                  biconfig.DeploymentStateService
	agentClientFactory                      bihttpagent.AgentClientFactory
	deploymentManagerFactory                bidepl.ManagerFactory
	deploymentManifestPath                  string
	deploymentVars                          boshtpl.Variables
	deploymentOp                            patch.Op
	releaseSetAndInstallationManifestParser ReleaseSetAndInstallationManifestParser
	deploymentManifestParser                DeploymentManifestParser
}

func (c *deploymentStateManager) StopDeployment(skipDrain bool, stage biui.Stage) (err error) {
	stateChanger := func(stage biui.Stage, directorID string, installationManifest biinstallmanifest.Manifest, update bideplmanifest.Update) (err error) {
		deployment, err := c.findDeploymend(directorID, installationManifest.Mbus, installationManifest.Cert.CA)

		if err != nil {
			return bosherr.WrapError(err, "Stopping environment")
		}

		return stage.PerformComplex("stopping deployment", func(stopEnvStage biui.Stage) error {
			return deployment.Stop(skipDrain, stopEnvStage)
		})
	}
	return c.executeStateChange(stage, stateChanger)
}

func (c *deploymentStateManager) StartDeployment(stage biui.Stage) (err error) {
	stateChanger := func(stage biui.Stage, directorID string, installationManifest biinstallmanifest.Manifest, update bideplmanifest.Update) (err error) {
		deployment, err := c.findDeploymend(directorID, installationManifest.Mbus, installationManifest.Cert.CA)

		if err != nil {
			return bosherr.WrapError(err, "Starting environment")
		}

		return stage.PerformComplex("starting deployment", func(startEnvStage biui.Stage) error {
			return deployment.Start(startEnvStage, update)
		})
	}
	return c.executeStateChange(stage, stateChanger)
}

func (c *deploymentStateManager) executeStateChange(stage biui.Stage, stateChanger func(biui.Stage, string, biinstallmanifest.Manifest, bideplmanifest.Update) error) (err error) {
	c.ui.BeginLinef("Deployment state: '%s'\n", c.deploymentStateService.Path())

	if !c.deploymentStateService.Exists() {
		c.ui.BeginLinef("No deployment state file found.\n")
		return nil
	}

	deploymentState, err := c.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment state")
	}

	var installationManifest biinstallmanifest.Manifest
	var releaseSetManifest birelsetmanifest.Manifest
	var update bideplmanifest.Update

	err = stage.PerformComplex("validating", func(stage biui.Stage) error {
		releaseSetManifest, installationManifest, err = c.releaseSetAndInstallationManifestParser.ReleaseSetAndInstallationManifest(c.deploymentManifestPath, c.deploymentVars, c.deploymentOp)
		if err != nil {
			return err
		}

		update, err = c.deploymentManifestParser.GetDeploymentManifestUpdate(c.deploymentManifestPath, c.deploymentVars, c.deploymentOp, releaseSetManifest, stage)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return stateChanger(stage, deploymentState.DirectorID, installationManifest, update)
}

func (c *deploymentStateManager) findDeploymend(directorID, installationMbus, caCert string) (bidepl.Deployment, error) {
	deploymentManager, err := c.deploymentManager(directorID, installationMbus, caCert)
	if err != nil {
		return nil, err
	}

	return c.findCurrentDeploymen(deploymentManager)

}

func (c *deploymentStateManager) findCurrentDeploymen(deploymentManager bidepl.Manager) (bidepl.Deployment, error) {
	c.logger.Debug(c.logTag, "Finding current deployment...")

	deployment, found, err := deploymentManager.FindCurrent()
	if err != nil {
		return nil, bosherr.WrapError(err, "Finding current deployment")
	}

	if !found {
		return nil, bosherr.WrapError(err, "No current deployment found...")
	}

	return deployment, err
}

func (c *deploymentStateManager) deploymentManager(directorID, installationMbus, caCert string) (bidepl.Manager, error) {

	c.logger.Debug(c.logTag, "Creating agent client...")

	agentClient, err := c.agentClientFactory.NewAgentClient(directorID, installationMbus, caCert)

	c.logger.Debug(c.logTag, "Creating deployment manager...")

	return c.deploymentManagerFactory.NewManager(nil, agentClient, nil), err
}
