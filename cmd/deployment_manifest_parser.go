package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cppforlife/go-patch/patch"

	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bidepltpl "github.com/cloudfoundry/bosh-cli/v7/deployment/template"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	birel "github.com/cloudfoundry/bosh-cli/v7/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type DeploymentManifestParser interface {
	GetDeploymentManifest(path string, vars boshtpl.Variables, op patch.Op, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage) (bideplmanifest.Manifest, string, error)
	GetDeploymentManifestUpdate(path string, vars boshtpl.Variables, op patch.Op, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage) (bideplmanifest.Update, error)
}

type deploymentManifestParser struct {
	deploymentParser    bideplmanifest.Parser
	deploymentValidator bideplmanifest.Validator
	releaseManager      birel.Manager
	templateFactory     bidepltpl.DeploymentTemplateFactory
}

func NewDeploymentManifestParser(
	deploymentParser bideplmanifest.Parser,
	deploymentValidator bideplmanifest.Validator,
	releaseManager birel.Manager,
	templateFactory bidepltpl.DeploymentTemplateFactory) DeploymentManifestParser {
	return deploymentManifestParser{
		deploymentParser:    deploymentParser,
		deploymentValidator: deploymentValidator,
		releaseManager:      releaseManager,
		templateFactory:     templateFactory,
	}
}

func (y deploymentManifestParser) GetDeploymentManifestUpdate(path string, vars boshtpl.Variables, op patch.Op, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage) (bideplmanifest.Update, error) {
	deploymentManifest, _, err := y.getManifest(path, vars, op, releaseSetManifest, stage, true)
	if err != nil {
		return bideplmanifest.Update{}, err
	}
	return deploymentManifest.Update, nil
}

func (y deploymentManifestParser) GetDeploymentManifest(path string, vars boshtpl.Variables, op patch.Op, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage) (bideplmanifest.Manifest, string, error) {
	return y.getManifest(path, vars, op, releaseSetManifest, stage, false)
}

func (y deploymentManifestParser) getManifest(path string, vars boshtpl.Variables, op patch.Op, releaseSetManifest birelsetmanifest.Manifest, stage biui.Stage, skipReleaseJobsValidation bool) (bideplmanifest.Manifest, string, error) {
	var deploymentManifest bideplmanifest.Manifest
	var manifestSHA string

	err := stage.Perform("Validating deployment manifest", func() error {
		var err error

		template, err := y.templateFactory.NewDeploymentTemplateFromPath(path)
		if err != nil {
			return bosherr.WrapErrorf(err, "Evaluating manifest")
		}

		interpolatedTemplate, err := template.Evaluate(vars, op)
		if err != nil {
			return bosherr.WrapErrorf(err, "Evaluating manifest '%s'", path)
		}

		manifestSHA = interpolatedTemplate.SHA()

		deploymentManifest, err = y.deploymentParser.Parse(interpolatedTemplate, path)
		if err != nil {
			return bosherr.WrapErrorf(err, "Parsing deployment manifest '%s'", path)
		}

		err = y.deploymentValidator.Validate(deploymentManifest, releaseSetManifest)
		if err != nil {
			return bosherr.WrapError(err, "Validating deployment manifest")
		}

		if !skipReleaseJobsValidation {
			err = y.deploymentValidator.ValidateReleaseJobs(deploymentManifest, y.releaseManager)
			if err != nil {
				return bosherr.WrapError(err, "Validating deployment jobs refer to jobs in release")
			}
		}

		return nil
	})
	if err != nil {
		return bideplmanifest.Manifest{}, "", err
	}

	return deploymentManifest, manifestSHA, nil
}
