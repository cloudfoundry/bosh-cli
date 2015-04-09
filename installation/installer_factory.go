package installation

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bideplrel "github.com/cloudfoundry/bosh-init/deployment/release"
	biindex "github.com/cloudfoundry/bosh-init/index"
	biinstallblob "github.com/cloudfoundry/bosh-init/installation/blob"
	biinstalljob "github.com/cloudfoundry/bosh-init/installation/job"
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	biinstallstate "github.com/cloudfoundry/bosh-init/installation/state"
	biregistry "github.com/cloudfoundry/bosh-init/registry"
	birelset "github.com/cloudfoundry/bosh-init/release/set"
	bistatejob "github.com/cloudfoundry/bosh-init/state/job"
	bistatepkg "github.com/cloudfoundry/bosh-init/state/pkg"
	bitemplate "github.com/cloudfoundry/bosh-init/templatescompiler"
	bierbrenderer "github.com/cloudfoundry/bosh-init/templatescompiler/erbrenderer"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type InstallerFactory interface {
	NewInstaller() (Installer, error)
}

type installerFactory struct {
	targetProvider        TargetProvider
	ui                    biui.UI
	fs                    boshsys.FileSystem
	runner                boshsys.CmdRunner
	extractor             boshcmd.Compressor
	releaseResolver       birelset.Resolver
	releaseJobResolver    bideplrel.JobResolver
	uuidGenerator         boshuuid.Generator
	registryServerManager biregistry.ServerManager
	logger                boshlog.Logger
	logTag                string
}

func NewInstallerFactory(
	targetProvider TargetProvider,
	ui biui.UI,
	fs boshsys.FileSystem,
	runner boshsys.CmdRunner,
	extractor boshcmd.Compressor,
	releaseResolver birelset.Resolver,
	releaseJobResolver bideplrel.JobResolver,
	uuidGenerator boshuuid.Generator,
	registryServerManager biregistry.ServerManager,
	logger boshlog.Logger,
) InstallerFactory {
	return &installerFactory{
		targetProvider:        targetProvider,
		ui:                    ui,
		fs:                    fs,
		runner:                runner,
		extractor:             extractor,
		releaseResolver:       releaseResolver,
		releaseJobResolver:    releaseJobResolver,
		uuidGenerator:         uuidGenerator,
		registryServerManager: registryServerManager,
		logger:                logger,
		logTag:                "installer",
	}
}

func (f *installerFactory) NewInstaller() (Installer, error) {

	target, err := f.targetProvider.NewTarget()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating installation target")
	}

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
	releaseJobResolver bideplrel.JobResolver

	stateBuilder          biinstallstate.Builder
	jobDependencyCompiler bistatejob.DependencyCompiler
	packageCompiler       bistatepkg.Compiler
	jobInstaller          biinstalljob.Installer
	templatesRepo         bitemplate.TemplatesRepo
	packageInstaller      biinstallpkg.Installer
	blobstore             boshblob.Blobstore
	blobExtractor         biinstallblob.Extractor
	compiledPackageRepo   bistatepkg.CompiledPackageRepo
}

func (c *installerFactoryContext) StateBuilder() biinstallstate.Builder {
	if c.stateBuilder != nil {
		return c.stateBuilder
	}

	erbRenderer := bierbrenderer.NewERBRenderer(c.fs, c.runner, c.logger)
	jobRenderer := bitemplate.NewJobRenderer(erbRenderer, c.fs, c.logger)
	jobListRenderer := bitemplate.NewJobListRenderer(jobRenderer, c.logger)

	c.stateBuilder = biinstallstate.NewBuilder(
		c.releaseJobResolver,
		c.JobDependencyCompiler(),
		jobListRenderer,
		c.extractor,
		c.Blobstore(),
		c.TemplatesRepo(),
	)
	return c.stateBuilder
}

func (c *installerFactoryContext) JobDependencyCompiler() bistatejob.DependencyCompiler {
	if c.jobDependencyCompiler != nil {
		return c.jobDependencyCompiler
	}

	c.jobDependencyCompiler = bistatejob.NewDependencyCompiler(
		c.PackageCompiler(),
		c.logger,
	)

	return c.jobDependencyCompiler
}

func (c *installerFactoryContext) PackageCompiler() bistatepkg.Compiler {
	if c.packageCompiler != nil {
		return c.packageCompiler
	}

	c.packageCompiler = biinstallpkg.NewPackageCompiler(
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

func (c *installerFactoryContext) JobInstaller() biinstalljob.Installer {
	if c.jobInstaller != nil {
		return c.jobInstaller
	}

	c.jobInstaller = biinstalljob.NewInstaller(
		c.fs,
		c.BlobExtractor(),
		c.TemplatesRepo(),
		c.target.JobsPath(),
	)
	return c.jobInstaller
}

func (c *installerFactoryContext) TemplatesRepo() bitemplate.TemplatesRepo {
	if c.templatesRepo != nil {
		return c.templatesRepo
	}

	templatesIndex := biindex.NewFileIndex(c.target.TemplatesIndexPath(), c.fs)
	c.templatesRepo = bitemplate.NewTemplatesRepo(templatesIndex)
	return c.templatesRepo
}

func (c *installerFactoryContext) PackageInstaller() biinstallpkg.Installer {
	if c.packageInstaller != nil {
		return c.packageInstaller
	}

	c.packageInstaller = biinstallpkg.NewPackageInstaller(c.BlobExtractor())
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

func (c *installerFactoryContext) BlobExtractor() biinstallblob.Extractor {
	if c.blobExtractor != nil {
		return c.blobExtractor
	}

	c.blobExtractor = biinstallblob.NewExtractor(c.fs, c.extractor, c.Blobstore(), c.logger)

	return c.blobExtractor
}

func (c *installerFactoryContext) CompiledPackageRepo() bistatepkg.CompiledPackageRepo {
	if c.compiledPackageRepo != nil {
		return c.compiledPackageRepo
	}

	compiledPackageIndex := biindex.NewFileIndex(c.target.CompiledPackagedIndexPath(), c.fs)
	c.compiledPackageRepo = bistatepkg.NewCompiledPackageRepo(compiledPackageIndex)

	return c.compiledPackageRepo
}
