package installation

import (
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	biregistry "github.com/cloudfoundry/bosh-init/registry"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Installer interface {
	InstallPackagesAndJobs(biinstallmanifest.Manifest, biui.Stage) (Installation, error)
	Cleanup(Installation) error
}

type installer struct {
	target                Target
	jobRenderer           JobRenderer
	jobResolver           JobResolver
	packageCompiler       PackageCompiler
	packagesPath          string
	packageInstaller      biinstallpkg.Installer
	jobInstaller          biinstalljob.Installer
	registryServerManager biregistry.ServerManager
	logger                boshlog.Logger
	logTag                string
}

func NewInstaller(
	target Target,
	jobRenderer JobRenderer,
	jobResolver JobResolver,
	packageCompiler PackageCompiler,
	packagesPath string,
	packageInstaller biinstallpkg.Installer,
	jobInstaller biinstalljob.Installer,
	registryServerManager biregistry.ServerManager,
	logger boshlog.Logger,
) Installer {
	return &installer{
		target:                target,
		jobRenderer:           jobRenderer,
		jobResolver:           jobResolver,
		packageCompiler:       packageCompiler,
		packagesPath:          packagesPath,
		packageInstaller:      packageInstaller,
		jobInstaller:          jobInstaller,
		registryServerManager: registryServerManager,
		logger:                logger,
		logTag:                "installer",
	}
}

func (i *installer) InstallPackagesAndJobs(manifest biinstallmanifest.Manifest, stage biui.Stage) (Installation, error) {
	i.logger.Info(i.logTag, "Installing CPI deployment '%s'", manifest.Name)
	i.logger.Debug(i.logTag, "Installing CPI deployment '%s' with manifest: %#v", manifest.Name, manifest)

	jobs, err := i.jobResolver.From(manifest)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving jobs from manifest")
	}

	compiledPackages, err := i.packageCompiler.For(jobs, i.packagesPath, stage)
	if err != nil {
		return nil, err
	}

	err = stage.Perform("Installing packages", func() error {
		return i.install(compiledPackages)
	})
	if err != nil {
		return nil, err
	}

	renderedJobRefs, err := i.jobRenderer.RenderAndUploadFrom(manifest, jobs, stage)
	renderedCPIJob := renderedJobRefs[0]
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

func (i *installer) Cleanup(installation Installation) error {
	return i.jobInstaller.Cleanup(installation.Job())
}

func (i *installer) install(compiledPackages []biinstallpkg.CompiledPackageRef) error {
	for _, compiledPackageRef := range compiledPackages {
		err := i.packageInstaller.Install(compiledPackageRef, i.packagesPath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Installing package '%s'", compiledPackageRef.Name)
		}
	}
	return nil
}
