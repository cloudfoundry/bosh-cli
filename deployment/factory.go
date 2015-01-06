package deployer

import (
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

type Factory interface {
	NewDeployment(bmdeplmanifest.Manifest) Deployment
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

func (f *factory) NewDeployment(manifest bmdeplmanifest.Manifest) Deployment {
	return NewDeployment(
		manifest,
		f.deployer,
	)
}
