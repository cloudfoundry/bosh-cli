package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bidepltpl "github.com/cloudfoundry/bosh-init/deployment/template"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	birel "github.com/cloudfoundry/bosh-init/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type DeploymentManifestParser struct {
	DeploymentParser    bideplmanifest.Parser
	DeploymentValidator bideplmanifest.Validator
	ReleaseManager      birel.Manager
	TemplateFactory     bidepltpl.DeploymentTemplateFactory
}

func (y DeploymentManifestParser) GetDeploymentManifest(path string, vars boshtpl.Variables, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage) (bideplmanifest.Manifest, string, error) {
	var deploymentManifest bideplmanifest.Manifest
	var manifestSHA string

	err := stage.Perform("Validating deployment manifest", func() error {
		var err error

		template, err := y.TemplateFactory.NewDeploymentTemplateFromPath(path)
		if err != nil {
			return bosherr.WrapErrorf(err, "Evaluating manifest")
		}

		interpolatedTemplate, err := template.Evaluate(vars)
		if err != nil {
			return bosherr.WrapErrorf(err, "Interpolating manifest '%s'", path)
		}

		manifestSHA = interpolatedTemplate.SHA()

		deploymentManifest, err = y.DeploymentParser.Parse(interpolatedTemplate, path)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", path)
		}

		err = y.DeploymentValidator.Validate(deploymentManifest, releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		err = y.DeploymentValidator.ValidateReleaseJobs(deploymentManifest, y.ReleaseManager)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment jobs refer to jobs in release")
		}

		return nil
	})
	if err != nil {
		return bideplmanifest.Manifest{}, "", err
	}

	return deploymentManifest, manifestSHA, nil
}
