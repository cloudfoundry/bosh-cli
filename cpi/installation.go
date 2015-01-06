package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type Installation interface {
	Install() (bmcloud.Cloud, error)
	StartJobs() error
	StopJobs() error
	Manifest() bminstallmanifest.Manifest
}

type installation struct {
	manifest              bminstallmanifest.Manifest
	registryServerManager bmregistry.ServerManager
	installer             Installer
	directorID            string

	release        bmrel.Release
	registryServer bmregistry.Server
}

func NewInstallation(
	manifest bminstallmanifest.Manifest,
	registryServerManager bmregistry.ServerManager,
	installer Installer,
	directorID string,
) Installation {
	return &installation{
		manifest:              manifest,
		registryServerManager: registryServerManager,
		installer:             installer,
		directorID:            directorID,
	}
}

func (i *installation) Manifest() bminstallmanifest.Manifest {
	return i.manifest
}

func (i *installation) Install() (bmcloud.Cloud, error) {
	return i.installer.Install(i.manifest, i.directorID)
}

func (i *installation) StartJobs() error {
	if !i.manifest.Registry.IsEmpty() {
		if i.registryServer != nil {
			return bosherr.Error("CPI jobs were already started")
		}
		registryServer, err := i.startRegistry()
		if err != nil {
			return bosherr.WrapError(err, "Starting registry")
		}
		i.registryServer = registryServer
	}
	return nil
}

func (i *installation) StopJobs() error {
	if !i.manifest.Registry.IsEmpty() {
		if i.registryServer == nil {
			return bosherr.Error("CPI jobs must be started before they can be stopped")
		}
		err := i.registryServer.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping registry")
		}
		i.registryServer = nil
	}
	return nil
}

func (i *installation) startRegistry() (bmregistry.Server, error) {
	config := i.manifest.Registry
	return i.registryServerManager.Start(config.Username, config.Password, config.Host, config.Port)
}
