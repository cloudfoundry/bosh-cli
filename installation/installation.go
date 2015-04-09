package installation

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biregistry "github.com/cloudfoundry/bosh-init/registry"
)

type Installation interface {
	Target() Target
	Job() biinstalljob.InstalledJob
	StartRegistry() error
	StopRegistry() error
}

type installation struct {
	target                Target
	job                   biinstalljob.InstalledJob
	manifest              biinstallmanifest.Manifest
	registryServerManager biregistry.ServerManager

	registryServer biregistry.Server
}

func NewInstallation(
	target Target,
	job biinstalljob.InstalledJob,
	manifest biinstallmanifest.Manifest,
	registryServerManager biregistry.ServerManager,
) Installation {
	return &installation{
		target:                target,
		job:                   job,
		manifest:              manifest,
		registryServerManager: registryServerManager,
	}
}

func (i *installation) Target() Target {
	return i.target
}

func (i *installation) Job() biinstalljob.InstalledJob {
	return i.job
}

func (i *installation) StartRegistry() error {
	if !i.manifest.Registry.IsEmpty() {
		if i.registryServer != nil {
			return bosherr.Error("Registry already started")
		}
		config := i.manifest.Registry
		registryServer, err := i.registryServerManager.Start(config.Username, config.Password, config.Host, config.Port)
		if err != nil {
			return bosherr.WrapError(err, "Starting registry")
		}
		i.registryServer = registryServer
	}
	return nil
}

func (i *installation) StopRegistry() error {
	if !i.manifest.Registry.IsEmpty() {
		if i.registryServer == nil {
			return bosherr.Error("Registry must be started before it can be stopped")
		}
		err := i.registryServer.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping registry")
		}
		i.registryServer = nil
	}
	return nil
}
