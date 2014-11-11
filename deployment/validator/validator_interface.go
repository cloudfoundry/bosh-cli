package validator

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type DeploymentValidator interface {
	Validate(deployment bmdepl.Deployment) error
}
