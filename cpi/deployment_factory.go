package cpi

import (
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
)

type DeploymentFactory interface {
	NewDeployment(bmmanifest.CPIDeploymentManifest) Deployment
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

func (f *deploymentFactory) NewDeployment(manifest bmmanifest.CPIDeploymentManifest) Deployment {
	return NewDeployment(
		manifest,
		f.registryServerManager,
		f.installer,
	)
}
