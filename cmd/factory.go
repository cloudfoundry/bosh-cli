package cmd

import (
	"errors"
	"time"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmcomp "github.com/cloudfoundry/bosh-micro-cli/cpi/compile"
	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/cpi/packages"
	bmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto"
	bmdeployer "github.com/cloudfoundry/bosh-micro-cli/deployer"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/deployer/blobstore"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmindex "github.com/cloudfoundry/bosh-micro-cli/index"
	bmrelvalidation "github.com/cloudfoundry/bosh-micro-cli/release/validation"
	bmtempcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmerbrenderer "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands                map[string](func() (Cmd, error))
	userConfig              bmconfig.UserConfig
	userConfigService       bmconfig.UserConfigService
	deploymentFile          bmconfig.DeploymentFile
	deploymentWorkspace     bmconfig.DeploymentWorkspace
	deploymentConfigService bmconfig.DeploymentConfigService
	fs                      boshsys.FileSystem
	ui                      bmui.UI
	logger                  boshlog.Logger
	uuidGenerator           boshuuid.Generator
	workspace               string
}

func NewFactory(
	userConfig bmconfig.UserConfig,
	userConfigService bmconfig.UserConfigService,
	fs boshsys.FileSystem,
	ui bmui.UI,
	logger boshlog.Logger,
	uuidGenerator boshuuid.Generator,
	workspace string,
) Factory {
	f := &factory{
		userConfig:        userConfig,
		userConfigService: userConfigService,
		fs:                fs,
		ui:                ui,
		logger:            logger,
		uuidGenerator:     uuidGenerator,
		workspace:         workspace,
	}
	f.loadDeploymentConfig()
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
		f.userConfig,
		f.userConfigService,
		f.deploymentFile,
		f.fs,
		f.uuidGenerator,
		f.logger,
	), nil
}

func (f *factory) createDeployCmd() (Cmd, error) {
	runner := boshsys.NewExecCmdRunner(f.logger)
	extractor := boshcmd.NewTarballCompressor(runner, f.fs)

	boshValidator := bmrelvalidation.NewBoshValidator(f.fs)
	cpiReleaseValidator := bmrelvalidation.NewCpiValidator()
	releaseValidator := bmrelvalidation.NewValidator(
		boshValidator,
		cpiReleaseValidator,
		f.ui,
	)

	compressor := boshcmd.NewTarballCompressor(runner, f.fs)
	indexFilePath := f.deploymentWorkspace.CompiledPackagedIndexPath()
	compiledPackageIndex := bmindex.NewFileIndex(indexFilePath, f.fs)
	compiledPackageRepo := bmpkgs.NewCompiledPackageRepo(compiledPackageIndex)

	options := map[string]interface{}{"blobstore_path": f.deploymentWorkspace.BlobstorePath()}
	blobstore := boshblob.NewSHA1VerifiableBlobstore(
		boshblob.NewLocalBlobstore(f.fs, f.uuidGenerator, options),
	)
	blobExtractor := bmcpiinstall.NewBlobExtractor(f.fs, compressor, blobstore, f.logger)
	packageInstaller := bmcpiinstall.NewPackageInstaller(compiledPackageRepo, blobExtractor)
	packageCompiler := bmcomp.NewPackageCompiler(
		runner,
		f.deploymentWorkspace.PackagesPath(),
		f.fs,
		compressor,
		blobstore,
		compiledPackageRepo,
		packageInstaller,
	)
	timeService := boshtime.NewConcreteService()
	eventFilters := []bmeventlog.EventFilter{
		bmeventlog.NewTimeFilter(timeService),
	}
	eventLogger := bmeventlog.NewEventLoggerWithFilters(f.ui, eventFilters)

	da := bmcomp.NewDependencyAnalysis()
	releasePackagesCompiler := bmcomp.NewReleasePackagesCompiler(
		da,
		packageCompiler,
		eventLogger,
		timeService,
	)

	cpiManifestParser := bmdepl.NewCpiDeploymentParser(f.fs)
	boshManifestParser := bmdepl.NewBoshDeploymentParser(f.fs, f.logger)
	boshDeploymentValidator := bmdeplval.NewBoshDeploymentValidator()
	erbRenderer := bmerbrenderer.NewERBRenderer(f.fs, runner, f.logger)
	jobRenderer := bmtempcomp.NewJobRenderer(erbRenderer, f.fs, f.logger)
	templatesIndex := bmindex.NewFileIndex(f.deploymentWorkspace.TemplatesIndexPath(), f.fs)
	templatesRepo := bmtempcomp.NewTemplatesRepo(templatesIndex)
	templatesCompiler := bmtempcomp.NewTemplatesCompiler(jobRenderer, compressor, blobstore, templatesRepo, f.fs, f.logger)
	releaseCompiler := bmcomp.NewReleaseCompiler(releasePackagesCompiler, templatesCompiler)
	jobInstaller := bmcpiinstall.NewJobInstaller(
		f.fs,
		packageInstaller,
		blobExtractor,
		templatesRepo,
		f.deploymentWorkspace.JobsPath(),
		f.deploymentWorkspace.PackagesPath(),
		eventLogger,
		timeService,
	)
	cloudFactory := bmcloud.NewFactory(f.fs, runner, f.deploymentWorkspace, f.logger)
	cpiInstaller := bmcpi.NewInstaller(
		f.ui,
		f.fs,
		extractor,
		releaseValidator,
		releaseCompiler,
		jobInstaller,
		cloudFactory,
		f.logger,
	)
	stemcellReader := bmstemcell.NewReader(compressor, f.fs)
	stemcellRepo := bmconfig.NewStemcellRepo(f.deploymentConfigService, f.uuidGenerator)
	stemcellExtractor := bmstemcell.NewExtractor(stemcellReader, f.fs)
	stemcellManagerFactory := bmstemcell.NewManagerFactory(stemcellRepo, eventLogger)

	agentClientFactory := bmagentclient.NewAgentClientFactory(f.deploymentWorkspace.DeploymentUUID(), 1*time.Second, f.logger)
	blobstoreFactory := bmblobstore.NewBlobstoreFactory(f.fs, f.logger)
	sha1Calculator := bmcrypto.NewSha1Calculator(f.fs)
	applySpecFactory := bmas.NewFactory()

	templatesSpecGenerator := bmas.NewTemplatesSpecGenerator(
		blobstoreFactory,
		compressor,
		jobRenderer,
		f.uuidGenerator,
		sha1Calculator,
		f.fs,
		f.logger,
	)

	vmRepo := bmconfig.NewVMRepo(f.deploymentConfigService)

	vmManagerFactory := bmvm.NewManagerFactory(
		vmRepo,
		agentClientFactory,
		f.deploymentConfigService,
		applySpecFactory,
		templatesSpecGenerator,
		f.fs,
		f.logger,
	)

	diskRepo := bmconfig.NewDiskRepo(f.deploymentConfigService, f.uuidGenerator)
	diskManagerFactory := bmdisk.NewManagerFactory(diskRepo, f.logger)
	registryServer := bmregistry.NewServer(f.logger)
	sshTunnelFactory := bmsshtunnel.NewFactory(f.logger)

	vmDeployer := bmvm.NewDeployer(vmManagerFactory, sshTunnelFactory, f.logger)
	deployer := bmdeployer.NewDeployer(
		vmDeployer,
		diskManagerFactory,
		registryServer,
		eventLogger,
		f.logger,
	)

	deploymentRepo := bmconfig.NewDeploymentRepo(f.deploymentConfigService)
	releaseRepo := bmconfig.NewReleaseRepo(f.deploymentConfigService, f.uuidGenerator)
	deploymentRecord := bmdeployer.NewDeploymentRecord(deploymentRepo, releaseRepo, stemcellRepo, sha1Calculator)

	return NewDeployCmd(
		f.ui,
		f.userConfig,
		f.fs,
		cpiManifestParser,
		boshManifestParser,
		boshDeploymentValidator,
		cpiInstaller,
		stemcellExtractor,
		stemcellManagerFactory,
		deploymentRecord,
		deployer,
		eventLogger,
		f.logger,
	), nil
}

func (f *factory) loadDeploymentConfig() error {
	f.deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(
		f.userConfig.DeploymentConfigFilePath(),
		f.fs,
		f.logger,
	)
	var err error
	f.deploymentFile, err = f.deploymentConfigService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading deployment config")
	}
	f.deploymentWorkspace = bmconfig.NewDeploymentWorkspace(f.workspace, f.deploymentFile.UUID)
	return nil
}
