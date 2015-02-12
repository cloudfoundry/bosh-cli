package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

func getDeploymentManifest(userConfig bmconfig.UserConfig, ui bmui.UI, fs boshsys.FileSystem) (manifestPath string, err error) {
	deploymentManifestPath := userConfig.DeploymentManifestPath

	if deploymentManifestPath == "" {
		ui.ErrorLinef("Deployment manifest not set")
		return "", bosherr.Error("Deployment manifest not set")
	}

	ui.PrintLinef("Deployment manifest: '%s'", deploymentManifestPath)

	if !fs.FileExists(deploymentManifestPath) {
		ui.ErrorLinef("Deployment manifest does not exist")
		return "", bosherr.Errorf("Deployment manifest does not exist at '%s'", deploymentManifestPath)
	}

	ui.PrintLinef("Deployment state: '%s'", userConfig.DeploymentConfigPath())

	return deploymentManifestPath, nil
}
