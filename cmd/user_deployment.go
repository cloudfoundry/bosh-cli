package cmd

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

func getDeploymentManifest(userConfig bmconfig.UserConfig, ui bmui.UI, fs boshsys.FileSystem) (manifestPath string, err error) {
	deploymentManifestPath := userConfig.DeploymentFile

	if deploymentManifestPath == "" {
		ui.Error("No deployment manifest set")
		return "", bosherr.Error("No deployment manifest set")
	}

	ui.Sayln(fmt.Sprintf("Current deployment manifest is '%s'", deploymentManifestPath))

	if !fs.FileExists(deploymentManifestPath) {
		ui.Error("Deployment manifest does not exist")
		return "", bosherr.Errorf("Deployment manifest does not exist at '%s'", deploymentManifestPath)
	}

	//TODO: Current deployment is `/vagrant/micro-aws.yml' (associated state file: `/vagrant/deployment.json')

	return deploymentManifestPath, nil
}
