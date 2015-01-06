package cpi

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcomp "github.com/cloudfoundry/bosh-micro-cli/cpi/compile"
	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/cpi/packages"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type InstallationFactory interface {
	NewInstallation(manifest bminstallmanifest.Manifest, deploymentID, directorID string) Installation
}

type installationFactory struct {
	releaseManager        bmrel.Manager
	registryServerManager bmregistry.ServerManager
	workspaceRootPath     string
	fs                    boshsys.FileSystem
	ui                    bmui.UI
	uuidGenerator         boshuuid.Generator
	eventLogger           bmeventlog.EventLogger
	timeService           boshtime.Service
	logger                boshlog.Logger
	logTag                string
}

func NewInstallationFactory(
	releaseManager bmrel.Manager,
	registryServerManager bmregistry.ServerManager,
	workspaceRootPath string,
	fs boshsys.FileSystem,
	ui bmui.UI,
	uuidGenerator boshuuid.Generator,
	eventLogger bmeventlog.EventLogger,
	timeService boshtime.Service,
	logger boshlog.Logger,
) InstallationFactory {
	return &installationFactory{
		releaseManager:        releaseManager,
		registryServerManager: registryServerManager,
		workspaceRootPath:     workspaceRootPath,
		fs:                    fs,
		ui:                    ui,
		uuidGenerator:         uuidGenerator,
		eventLogger:           eventLogger,
		timeService:           timeService,
		logger:                logger,
		logTag:                "installationFactory",
	}
}

func (f *installationFactory) NewInstallation(manifest bminstallmanifest.Manifest, deploymentID, directorID string) Installation {
	return NewInstallation(
		manifest,
		f.registryServerManager,
		f.newCPIInstaller(deploymentID),
		directorID,
	)
}

func (f *installationFactory) newCPIInstaller(deploymentID string) Installer {
	deploymentWorkspace := bmconfig.NewDeploymentWorkspace(f.workspaceRootPath, deploymentID)

	runner := boshsys.NewExecCmdRunner(f.logger)
	extractor := boshcmd.NewTarballCompressor(runner, f.fs)

	compressor := boshcmd.NewTarballCompressor(runner, f.fs)
	indexFilePath := deploymentWorkspace.CompiledPackagedIndexPath()
	compiledPackageIndex := bmindex.NewFileIndex(indexFilePath, f.fs)
	compiledPackageRepo := bmpkgs.NewCompiledPackageRepo(compiledPackageIndex)

	options := map[string]interface{}{"blobstore_path": deploymentWorkspace.BlobstorePath()}
	localBlobstore := boshblob.NewLocalBlobstore(f.fs, f.uuidGenerator, options)
	blobstore := boshblob.NewSHA1VerifiableBlobstore(localBlobstore)
	blobExtractor := bmcpiinstall.NewBlobExtractor(f.fs, compressor, blobstore, f.logger)
	packageInstaller := bmcpiinstall.NewPackageInstaller(compiledPackageRepo, blobExtractor)
	packageCompiler := bmcomp.NewPackageCompiler(
		runner,
		deploymentWorkspace.PackagesPath(),
		f.fs,
		compressor,
		blobstore,
		compiledPackageRepo,
		packageInstaller,
	)

	da := bmcomp.NewDependencyAnalysis()
	releasePackagesCompiler := bmcomp.NewReleasePackagesCompiler(
		da,
		packageCompiler,
		f.eventLogger,
		f.timeService,
	)

	erbRenderer := bmerbrenderer.NewERBRenderer(f.fs, runner, f.logger)
	jobRenderer := bmtempcomp.NewJobRenderer(erbRenderer, f.fs, f.logger)
	templatesIndex := bmindex.NewFileIndex(deploymentWorkspace.TemplatesIndexPath(), f.fs)
	templatesRepo := bmtempcomp.NewTemplatesRepo(templatesIndex)
	templatesCompiler := bmtempcomp.NewTemplatesCompiler(jobRenderer, compressor, blobstore, templatesRepo, f.fs, f.logger)
	releaseCompiler := bmcomp.NewReleaseCompiler(releasePackagesCompiler, templatesCompiler, f.logger)
	jobInstaller := bmcpiinstall.NewJobInstaller(
		f.fs,
		packageInstaller,
		blobExtractor,
		templatesRepo,
		deploymentWorkspace.JobsPath(),
		deploymentWorkspace.PackagesPath(),
		f.eventLogger,
		f.timeService,
	)
	cloudFactory := bmcloud.NewFactory(f.fs, runner, deploymentWorkspace, f.logger)

	return NewInstaller(
		f.ui,
		f.fs,
		extractor,
		f.releaseManager,
		releaseCompiler,
		jobInstaller,
		cloudFactory,
		f.logger,
	)
}
