package deployer

import (
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

type Factory interface {
	NewDeployment(bmmanifest.Manifest) Deployment
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

func (f *factory) NewDeployment(manifest bmmanifest.Manifest) Deployment {
	return NewDeployment(
		manifest,
		f.deployer,
	)
}
