package cmd

import (
	"errors"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmcomp "github.com/cloudfoundry/bosh-micro-cli/compile"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdeploy "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/erbrenderer"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/install"
	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/packages"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands      map[string](func() (Cmd, error))
	config        bmconfig.Config
	configService bmconfig.Service
	fileSystem    boshsys.FileSystem
	ui            bmui.UI
	logger        boshlog.Logger
	uuidGenerator boshuuid.Generator
}

func NewFactory(
	config bmconfig.Config,
	configService bmconfig.Service,
	fileSystem boshsys.FileSystem,
	ui bmui.UI,
	logger boshlog.Logger,
	uuidGenerator boshuuid.Generator,

) Factory {
	f := &factory{
		config:        config,
		configService: configService,
		fileSystem:    fileSystem,
		ui:            ui,
		logger:        logger,
		uuidGenerator: uuidGenerator,
	}
	f.commands = map[string](func() (Cmd, error)){
		"deployment": f.createDeploymentCmd,
		"deploy":     f.createDeployCmd,
	}
	return f
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name]()
}

func (f *factory) createDeploymentCmd() (Cmd, error) {
	return NewDeploymentCmd(
		f.ui,
		f.config,
		f.configService,
		f.fileSystem,
		f.uuidGenerator,
		f.logger,
	), nil
}

func (f *factory) createDeployCmd() (Cmd, error) {
	runner := boshsys.NewExecCmdRunner(f.logger)
	extractor := boshcmd.NewTarballCompressor(runner, f.fileSystem)

	boshValidator := bmrelvalidation.NewBoshValidator(f.fileSystem)
	cpiReleaseValidator := bmrelvalidation.NewCpiValidator()
	releaseValidator := bmrelvalidation.NewValidator(
		boshValidator,
		cpiReleaseValidator,
		f.ui,
	)

	compressor := boshcmd.NewTarballCompressor(runner, f.fileSystem)
	indexFilePath := f.config.CompiledPackagedIndexPath()
	compiledPackageIndex := bmindex.NewFileIndex(indexFilePath, f.fileSystem)
	compiledPackageRepo := bmpkgs.NewCompiledPackageRepo(compiledPackageIndex)

	options := map[string]interface{}{"blobstore_path": f.config.BlobstorePath()}
	blobstore := boshblob.NewSHA1VerifiableBlobstore(
		boshblob.NewLocalBlobstore(f.fileSystem, f.uuidGenerator, options),
	)
	blobExtractor := bminstall.NewBlobExtractor(f.fileSystem, compressor, blobstore, f.logger)
	packageInstaller := bminstall.NewPackageInstaller(compiledPackageRepo, blobExtractor)
	packageCompiler := bmcomp.NewPackageCompiler(
		runner,
		f.config.PackagesPath(),
		f.fileSystem,
		compressor,
		blobstore,
		compiledPackageRepo,
		packageInstaller,
	)
	eventLogger := bmlog.NewEventLogger(f.ui)

	da := bmcomp.NewDependencyAnalysis()
	timeService := boshtime.NewConcreteService()
	releasePackagesCompiler := bmcomp.NewReleasePackagesCompiler(
		da,
		packageCompiler,
		eventLogger,
		timeService,
	)

	manifestParser := bmdepl.NewCpiDeploymentParser(f.fileSystem)
	erbrenderer := bmerbrenderer.NewERBRenderer(f.fileSystem, runner, f.logger)
	templatesIndex := bmindex.NewFileIndex(f.config.TemplatesIndexPath(), f.fileSystem)
	templatesRepo := bmtempcomp.NewTemplatesRepo(templatesIndex)
	templatesCompiler := bmtempcomp.NewTemplatesCompiler(erbrenderer, compressor, blobstore, templatesRepo, f.fileSystem, f.logger)
	releaseCompiler := bmcomp.NewReleaseCompiler(releasePackagesCompiler, templatesCompiler)
	jobInstaller := bminstall.NewJobInstaller(
		f.fileSystem,
		packageInstaller,
		blobExtractor,
		templatesRepo,
		f.config.PackagesPath(),
		f.config.JobsPath(),
		eventLogger,
		timeService,
	)
	cloudFactory := bmcloud.NewFactory(f.fileSystem, runner, f.config.DeploymentUUID, f.logger)
	cpiDeployer := bmdeploy.NewCpiDeployer(
		f.ui,
		f.fileSystem,
		extractor,
		releaseValidator,
		releaseCompiler,
		jobInstaller,
		cloudFactory,
		f.logger,
	)
	stemcellReader := bmstemcell.NewReader(compressor, f.fileSystem)
	repo := bmstemcell.NewRepo(f.configService)
	stemcellManagerFactory := bmstemcell.NewManagerFactory(f.fileSystem, stemcellReader, repo)

	return NewDeployCmd(
		f.ui,
		f.config,
		f.fileSystem,
		manifestParser,
		cpiDeployer,
		stemcellManagerFactory,
		f.logger,
	), nil
}
