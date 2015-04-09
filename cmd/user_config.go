package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	biconfig "github.com/cloudfoundry/bosh-init/config"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

func getDeploymentManifest(userConfig biconfig.UserConfig, ui biui.UI, fs boshsys.FileSystem) (manifestPath string, err error) {
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
