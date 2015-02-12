package installation

import (
	"path/filepath"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bminstallblob "github.com/cloudfoundry/bosh-micro-cli/installation/blob"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bminstallstate "github.com/cloudfoundry/bosh-micro-cli/installation/state"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmstatepkg "github.com/cloudfoundry/bosh-micro-cli/state/pkg"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
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
	releaseJobResolver      bmdeplrel.JobResolver
	workspaceRootPath       string
	uuidGenerator           boshuuid.Generator
	registryServerManager   bmregistry.ServerManager
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
	releaseJobResolver bmdeplrel.JobResolver,
	workspaceRootPath string,
	uuidGenerator boshuuid.Generator,
	registryServerManager bmregistry.ServerManager,
	logger boshlog.Logger,
) InstallerFactory {
	return &installerFactory{
		ui:                      ui,
		fs:                      fs,
		runner:                  runner,
		extractor:               extractor,
		deploymentConfigService: deploymentConfigService,
		releaseResolver:         releaseResolver,
		releaseJobResolver:      releaseJobResolver,
		workspaceRootPath:       workspaceRootPath,
		uuidGenerator:           uuidGenerator,
		registryServerManager:   registryServerManager,
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
		target:             target,
		fs:                 f.fs,
		runner:             f.runner,
		logger:             f.logger,
		extractor:          f.extractor,
		uuidGenerator:      f.uuidGenerator,
		releaseJobResolver: f.releaseJobResolver,
	}

	return NewInstaller(
		target,
		f.fs,
		context.StateBuilder(),
		target.PackagesPath(),
		context.PackageInstaller(),
		context.JobInstaller(),
		f.registryServerManager,
		f.logger,
	), nil
}

type installerFactoryContext struct {
	target             Target
	fs                 boshsys.FileSystem
	runner             boshsys.CmdRunner
	logger             boshlog.Logger
	extractor          boshcmd.Compressor
	uuidGenerator      boshuuid.Generator
	releaseJobResolver bmdeplrel.JobResolver

	stateBuilder        bminstallstate.Builder
	packageCompiler     bmstatepkg.Compiler
	jobInstaller        bminstalljob.Installer
	templatesRepo       bmtemplate.TemplatesRepo
	packageInstaller    bminstallpkg.Installer
	blobstore           boshblob.Blobstore
	blobExtractor       bminstallblob.Extractor
	compiledPackageRepo bmstatepkg.CompiledPackageRepo
}

func (c *installerFactoryContext) StateBuilder() bminstallstate.Builder {
	if c.stateBuilder != nil {
		return c.stateBuilder
	}

	erbRenderer := bmerbrenderer.NewERBRenderer(c.fs, c.runner, c.logger)
	jobRenderer := bmtemplate.NewJobRenderer(erbRenderer, c.fs, c.logger)
	jobListRenderer := bmtemplate.NewJobListRenderer(jobRenderer, c.logger)

	c.stateBuilder = bminstallstate.NewBuilder(
		c.releaseJobResolver,
		c.PackageCompiler(),
		jobListRenderer,
		c.extractor,
		c.Blobstore(),
		c.TemplatesRepo(),
	)
	return c.stateBuilder
}

func (c *installerFactoryContext) PackageCompiler() bmstatepkg.Compiler {
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
		c.logger,
	)

	return c.packageCompiler
}

func (c *installerFactoryContext) JobInstaller() bminstalljob.Installer {
	if c.jobInstaller != nil {
		return c.jobInstaller
	}

	c.jobInstaller = bminstalljob.NewInstaller(
		c.fs,
		c.BlobExtractor(),
		c.TemplatesRepo(),
		c.target.JobsPath(),
	)
	return c.jobInstaller
}

func (c *installerFactoryContext) TemplatesRepo() bmtemplate.TemplatesRepo {
	if c.templatesRepo != nil {
		return c.templatesRepo
	}

	templatesIndex := bmindex.NewFileIndex(c.target.TemplatesIndexPath(), c.fs)
	c.templatesRepo = bmtemplate.NewTemplatesRepo(templatesIndex)
	return c.templatesRepo
}

func (c *installerFactoryContext) PackageInstaller() bminstallpkg.Installer {
	if c.packageInstaller != nil {
		return c.packageInstaller
	}

	c.packageInstaller = bminstallpkg.NewPackageInstaller(c.BlobExtractor())
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

func (c *installerFactoryContext) CompiledPackageRepo() bmstatepkg.CompiledPackageRepo {
	if c.compiledPackageRepo != nil {
		return c.compiledPackageRepo
	}

	compiledPackageIndex := bmindex.NewFileIndex(c.target.CompiledPackagedIndexPath(), c.fs)
	c.compiledPackageRepo = bmstatepkg.NewCompiledPackageRepo(compiledPackageIndex)

	return c.compiledPackageRepo
}
