package deployer

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type Factory interface {
	NewDeployment(bmdepl.Manifest) Deployment
}

type factory struct {
	deployer Deployer
}

func NewFactory(
	deployer Deployer,
) Factory {
	return &factory{
		deployer: deployer,
	}
}

func (f *factory) NewDeployment(manifest bmdepl.Manifest) Deployment {
	return NewDeployment(
		manifest,
		f.deployer,
	)
}
