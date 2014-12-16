package validator

import (
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

type DeploymentValidator interface {
	Validate(deploymentManifest bmmanifest.Manifest) error
}
