package installation

import (
	"os"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bminstallstate "github.com/cloudfoundry/bosh-micro-cli/installation/state"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
)

type Installer interface {
	Install(bminstallmanifest.Manifest, bmeventlog.Stage) (Installation, error)
}

type installer struct {
	target                Target
	fs                    boshsys.FileSystem
	stateBuilder          bminstallstate.Builder
	packagesPath          string
	packageInstaller      bminstallpkg.PackageInstaller
	jobInstaller          bminstalljob.Installer
	registryServerManager bmregistry.ServerManager
	logger                boshlog.Logger
	logTag                string
}

func NewInstaller(
	target Target,
	fs boshsys.FileSystem,
	stateBuilder bminstallstate.Builder,
	packagesPath string,
	packageInstaller bminstallpkg.PackageInstaller,
	jobInstaller bminstalljob.Installer,
	registryServerManager bmregistry.ServerManager,
	logger boshlog.Logger,
) Installer {
	return &installer{
		target:                target,
		fs:                    fs,
		stateBuilder:          stateBuilder,
		packagesPath:          packagesPath,
		packageInstaller:      packageInstaller,
		jobInstaller:          jobInstaller,
		registryServerManager: registryServerManager,
		logger:                logger,
		logTag:                "installer",
	}
}

func (i *installer) Install(manifest bminstallmanifest.Manifest, stage bmeventlog.Stage) (Installation, error) {
	i.logger.Info(i.logTag, "Installing CPI deployment '%s'", manifest.Name)
	i.logger.Debug(i.logTag, "Installing CPI deployment '%s' with manifest: %#v", manifest.Name, manifest)

	state, err := i.stateBuilder.Build(manifest, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Building installation state")
	}

	err = stage.PerformStep("Installing packages", func() error {
		err = i.fs.MkdirAll(i.packagesPath, os.ModePerm)
		if err != nil {
			return bosherr.WrapErrorf(err, "Creating packages directory '%s'", i.packagesPath)
		}

		for _, compiledPackageRef := range state.CompiledPackages() {
			err = i.packageInstaller.Install(compiledPackageRef, i.packagesPath)
			if err != nil {
				return bosherr.WrapErrorf(err, "Installing package '%s'", compiledPackageRef.Name)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	renderedCPIJob := state.RenderedCPIJob()
	installedJob, err := i.jobInstaller.Install(renderedCPIJob, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Installing job '%s' for CPI release", renderedCPIJob.Name)
	}

	return NewInstallation(
		i.target,
		installedJob,
		manifest,
		i.registryServerManager,
	), nil
}
