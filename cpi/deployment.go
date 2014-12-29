package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type Deployment interface {
	ExtractRelease(releaseTarballPath string) (bmrel.Release, error)
	Install() (bmcloud.Cloud, error)
	StartJobs() error
	StopJobs() error
	Manifest() bmmanifest.CPIDeploymentManifest
}

type deployment struct {
	manifest              bmmanifest.CPIDeploymentManifest
	registryServerManager bmregistry.ServerManager
	installer             Installer
	directorID            string

	release        bmrel.Release
	registryServer bmregistry.Server
}

func NewDeployment(
	manifest bmmanifest.CPIDeploymentManifest,
	registryServerManager bmregistry.ServerManager,
	installer Installer,
	directorID string,
) Deployment {
	return &deployment{
		manifest:              manifest,
		registryServerManager: registryServerManager,
		installer:             installer,
		directorID:            directorID,
	}
}

func (d *deployment) Manifest() bmmanifest.CPIDeploymentManifest {
	return d.manifest
}

func (d *deployment) ExtractRelease(releaseTarballPath string) (bmrel.Release, error) {
	release, err := d.installer.Extract(releaseTarballPath)
	if err != nil {
		return release, err
	}
	if err == nil {
		// store extracted release for use by installer (may be deleted later)
		d.release = release
	}
	return release, err
}

func (d *deployment) Install() (bmcloud.Cloud, error) {
	if d.release == nil {
		return nil, bosherr.Error("CPI release must be extracted before it can be installed")
	}
	if !d.release.Exists() {
		return nil, bosherr.Errorf("Extracted CPI release not found")
	}
	return d.installer.Install(d.manifest, d.release, d.directorID)
}

func (d *deployment) StartJobs() error {
	if !d.manifest.Registry.IsEmpty() {
		if d.registryServer != nil {
			return bosherr.Error("CPI jobs were already started")
		}
		registryServer, err := d.startRegistry()
		if err != nil {
			return bosherr.WrapError(err, "Starting registry")
		}
		d.registryServer = registryServer
	}
	return nil
}

func (d *deployment) StopJobs() error {
	if !d.manifest.Registry.IsEmpty() {
		if d.registryServer == nil {
			return bosherr.Error("CPI jobs must be started before they can be stopped")
		}
		err := d.registryServer.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping registry")
		}
		d.registryServer = nil
	}
	return nil
}

func (d *deployment) startRegistry() (bmregistry.Server, error) {
	config := d.manifest.Registry
	return d.registryServerManager.Start(config.Username, config.Password, config.Host, config.Port)
}
