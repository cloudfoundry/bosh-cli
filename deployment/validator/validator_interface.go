package validator

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type DeploymentValidator interface {
	Validate(deploymentManifest bmdepl.Manifest) error
}
