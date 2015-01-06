package validator

import (
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

type DeploymentValidator interface {
	Validate(deploymentManifest bmdeplmanifest.Manifest) error
}
