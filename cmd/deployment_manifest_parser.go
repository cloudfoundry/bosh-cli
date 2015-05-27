package cmd

import (
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	birel "github.com/cloudfoundry/bosh-init/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	biui "github.com/cloudfoundry/bosh-init/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type deploymentParser struct {
	deploymentParser    bideplmanifest.Parser
	deploymentValidator bideplmanifest.Validator
	releaseManager      birel.Manager
}

func (y deploymentParser) GetDeploymentManifest(deploymentManifestPath string, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage) (bideplmanifest.Manifest, error) {
	var deploymentManifest bideplmanifest.Manifest
	err := stage.Perform("Validating deployment manifest", func() error {
		var err error
		deploymentManifest, err = y.deploymentParser.Parse(deploymentManifestPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", deploymentManifestPath)
		}

		err = y.deploymentValidator.Validate(deploymentManifest, releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		err = y.deploymentValidator.ValidateReleaseJobs(deploymentManifest, y.releaseManager)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment jobs refer to jobs in release")
		}

		return nil
	})
	if err != nil {
		return bideplmanifest.Manifest{}, err
	}

	return deploymentManifest, nil
}
