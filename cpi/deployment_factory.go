package cpi

import (
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
)

type DeploymentFactory interface {
	NewDeployment(bmdepl.CPIDeploymentManifest) Deployment
}

type deploymentFactory struct {
	registryServerManager bmregistry.ServerManager
	installer             Installer
}

func NewDeploymentFactory(
	registryServerManager bmregistry.ServerManager,
	installer Installer,
) DeploymentFactory {
	return &deploymentFactory{
		registryServerManager: registryServerManager,
		installer:             installer,
	}
}

func (f *deploymentFactory) NewDeployment(manifest bmdepl.CPIDeploymentManifest) Deployment {
	return NewDeployment(
		manifest,
		f.registryServerManager,
		f.installer,
	)
}
