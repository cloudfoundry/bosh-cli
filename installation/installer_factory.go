package installation

import (
	"path/filepath"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bminstallblob "github.com/cloudfoundry/bosh-micro-cli/installation/blob"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type InstallerFactory interface {
	NewInstaller() (Installer, error)
}

type installerFactory struct {
	ui                      bmui.UI
	fs                      boshsys.FileSystem
	runner                  boshsys.CmdRunner
	extractor               boshcmd.Compressor
	deploymentConfigService bmconfig.DeploymentConfigService
	releaseResolver         bmrelset.Resolver
	workspaceRootPath       string
	uuidGenerator           boshuuid.Generator
	timeService             boshtime.Service
	registryServerManager   bmregistry.ServerManager
	eventLogger             bmeventlog.EventLogger
	logger                  boshlog.Logger
	logTag                  string
}

func NewInstallerFactory(
	ui bmui.UI,
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	extractor boshcmd.Compressor,
	deploymentConfigService bmconfig.DeploymentConfigService,
	releaseResolver bmrelset.Resolver,
	workspaceRootPath string,
	uuidGenerator boshuuid.Generator,
	timeService boshtime.Service,
	registryServerManager bmregistry.ServerManager,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) InstallerFactory {
	return &installerFactory{
		ui:                      ui,
		fs:                      fs,
		runner:                  runner,
		extractor:               extractor,
		deploymentConfigService: deploymentConfigService,
		releaseResolver:         releaseResolver,
		workspaceRootPath:       workspaceRootPath,
		uuidGenerator:           uuidGenerator,
		timeService:             timeService,
		registryServerManager:   registryServerManager,
		eventLogger:             eventLogger,
		logger:                  logger,
		logTag:                  "installer",
	}
}

func (f *installerFactory) NewInstaller() (Installer, error) {
	deploymentConfig, err := f.deploymentConfigService.Load()
	if err != nil {
		return nil, bosherr.WrapError(err, "Loading deployment config")
	}

	installationID := deploymentConfig.InstallationID
	if installationID != "" {
		installationID, err = f.uuidGenerator.Generate()
		if err != nil {
			return nil, bosherr.WrapError(err, "Generating installation ID")
		}

		deploymentConfig.InstallationID = installationID
		err := f.deploymentConfigService.Save(deploymentConfig)
		if err != nil {
			return nil, bosherr.WrapError(err, "Saving deployment config")
		}
	}

	target := NewTarget(filepath.Join(f.workspaceRootPath, installationID))

	context := &installerFactoryContext{
		target:        target,
		eventLogger:   f.eventLogger,
		timeService:   f.timeService,
		fs:            f.fs,
		runner:        f.runner,
		logger:        f.logger,
		extractor:     f.extractor,
		uuidGenerator: f.uuidGenerator,
	}

	return NewInstaller(
		target,
		f.ui,
		f.releaseResolver,
		context.ReleaseCompiler(),
		context.JobInstaller(),
		f.registryServerManager,
		f.logger,
	), nil
}

type installerFactoryContext struct {
	target        Target
	eventLogger   bmeventlog.EventLogger
	timeService   boshtime.Service
	fs            boshsys.FileSystem
	runner        boshsys.CmdRunner
	logger        boshlog.Logger
	extractor     boshcmd.Compressor
	uuidGenerator boshuuid.Generator

	releaseCompiler     bminstallpkg.ReleaseCompiler
	packageCompiler     bminstallpkg.PackageCompiler
	jobInstaller        bminstalljob.Installer
	templatesRepo       bmtempcomp.TemplatesRepo
	packageInstaller    bminstallpkg.PackageInstaller
	blobstore           boshblob.Blobstore
	blobExtractor       bminstallblob.Extractor
	compiledPackageRepo bminstallpkg.CompiledPackageRepo
}

func (c *installerFactoryContext) ReleaseCompiler() bminstallpkg.ReleaseCompiler {
	if c.releaseCompiler != nil {
		return c.releaseCompiler
	}

	releasePackagesCompiler := bminstallpkg.NewReleasePackagesCompiler(
		c.PackageCompiler(),
		c.eventLogger,
		c.timeService,
	)

	erbRenderer := bmerbrenderer.NewERBRenderer(c.fs, c.runner, c.logger)
	jobRenderer := bmtempcomp.NewJobRenderer(erbRenderer, c.fs, c.logger)

	templatesCompiler := bmtempcomp.NewTemplatesCompiler(jobRenderer, c.extractor, c.Blobstore(), c.TemplatesRepo(), c.fs, c.logger)
	c.releaseCompiler = bminstallpkg.NewReleaseCompiler(releasePackagesCompiler, templatesCompiler, c.logger)
	return c.releaseCompiler
}

func (c *installerFactoryContext) PackageCompiler() bminstallpkg.PackageCompiler {
	if c.packageCompiler != nil {
		return c.packageCompiler
	}

	c.packageCompiler = bminstallpkg.NewPackageCompiler(
		c.runner,
		c.target.PackagesPath(),
		c.fs,
		c.extractor,
		c.Blobstore(),
		c.CompiledPackageRepo(),
		c.PackageInstaller(),
	)

	return c.packageCompiler
}

func (c *installerFactoryContext) JobInstaller() bminstalljob.Installer {
	if c.jobInstaller != nil {
		return c.jobInstaller
	}

	c.jobInstaller = bminstalljob.NewInstaller(
		c.fs,
		c.PackageInstaller(),
		c.BlobExtractor(),
		c.TemplatesRepo(),
		c.target.JobsPath(),
		c.target.PackagesPath(),
		c.timeService,
	)
	return c.jobInstaller
}

func (c *installerFactoryContext) TemplatesRepo() bmtempcomp.TemplatesRepo {
	if c.templatesRepo != nil {
		return c.templatesRepo
	}

	templatesIndex := bmindex.NewFileIndex(c.target.TemplatesIndexPath(), c.fs)
	c.templatesRepo = bmtempcomp.NewTemplatesRepo(templatesIndex)
	return c.templatesRepo
}

func (c *installerFactoryContext) PackageInstaller() bminstallpkg.PackageInstaller {
	if c.packageInstaller != nil {
		return c.packageInstaller
	}

	c.packageInstaller = bminstallpkg.NewPackageInstaller(c.CompiledPackageRepo(), c.BlobExtractor())
	return c.packageInstaller
}

func (c *installerFactoryContext) Blobstore() boshblob.Blobstore {
	if c.blobstore != nil {
		return c.blobstore
	}

	options := map[string]interface{}{"blobstore_path": c.target.BlobstorePath()}
	localBlobstore := boshblob.NewLocalBlobstore(c.fs, c.uuidGenerator, options)
	c.blobstore = boshblob.NewSHA1VerifiableBlobstore(localBlobstore)

	return c.blobstore
}

func (c *installerFactoryContext) BlobExtractor() bminstallblob.Extractor {
	if c.blobExtractor != nil {
		return c.blobExtractor
	}

	c.blobExtractor = bminstallblob.NewExtractor(c.fs, c.extractor, c.Blobstore(), c.logger)

	return c.blobExtractor
}

func (c *installerFactoryContext) CompiledPackageRepo() bminstallpkg.CompiledPackageRepo {
	if c.compiledPackageRepo != nil {
		return c.compiledPackageRepo
	}

	indexFilePath := c.target.CompiledPackagedIndexPath()
	compiledPackageIndex := bmindex.NewFileIndex(indexFilePath, c.fs)
	c.compiledPackageRepo = bminstallpkg.NewCompiledPackageRepo(compiledPackageIndex)

	return c.compiledPackageRepo
}
