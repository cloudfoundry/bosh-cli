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
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
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
	runner                  boshsys.CmdRunner
	compressor              boshcmd.Compressor
	vmManagerFactory        bmvm.ManagerFactory
	vmRepo                  bmconfig.VMRepo
	stemcellRepo            bmconfig.StemcellRepo
	diskRepo                bmconfig.DiskRepo
	registryServerFactory   bmregistry.ServerFactory
	sshTunnelFactory        bmsshtunnel.Factory
	diskDeployer            bminstance.DiskDeployer
	diskManagerFactory      bmdisk.ManagerFactory
	instanceManagerFactory  bminstance.ManagerFactory
	stemcellManagerFactory  bmstemcell.ManagerFactory
	eventLogger             bmeventlog.EventLogger
	timeService             boshtime.Service
	cpiInstaller            bmcpi.Installer
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
		"delete":     f.createDeleteCmd,
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
	deploymentParser := bmdepl.NewParser(f.fs, f.logger)

	boshDeploymentValidator := bmdeplval.NewBoshDeploymentValidator()

	stemcellReader := bmstemcell.NewReader(f.loadCompressor(), f.fs)
	stemcellExtractor := bmstemcell.NewExtractor(stemcellReader, f.fs)

	deploymentRepo := bmconfig.NewDeploymentRepo(f.deploymentConfigService)
	releaseRepo := bmconfig.NewReleaseRepo(f.deploymentConfigService, f.uuidGenerator)
	sha1Calculator := bmcrypto.NewSha1Calculator(f.fs)
	deploymentRecord := bmdeployer.NewDeploymentRecord(deploymentRepo, releaseRepo, f.loadStemcellRepo(), sha1Calculator)

	deployer := bmdeployer.NewDeployer(
		f.loadStemcellManagerFactory(),
		f.loadVMManagerFactory(),
		f.loadSSHTunnelFactory(),
		f.loadDiskDeployer(),
		f.loadRegistryServerFactory(),
		f.loadEventLogger(),
		f.logger,
	)

	return NewDeployCmd(
		f.ui,
		f.userConfig,
		f.fs,
		deploymentParser,
		boshDeploymentValidator,
		f.loadCPIInstaller(),
		stemcellExtractor,
		deploymentRecord,
		deployer,
		f.loadEventLogger(),
		f.logger,
	), nil
}

func (f *factory) createDeleteCmd() (Cmd, error) {
	deploymentParser := bmdepl.NewParser(f.fs, f.logger)

	agentClientFactory := bmagentclient.NewAgentClientFactory(f.deploymentWorkspace.DeploymentUUID(), 1*time.Second, f.logger)

	return NewDeleteCmd(
		f.ui,
		f.userConfig,
		f.fs,
		deploymentParser,
		f.loadCPIInstaller(),
		f.loadVMManagerFactory(),
		f.loadInstanceManagerFactory(),
		f.loadDiskManagerFactory(),
		f.loadStemcellManagerFactory(),
		agentClientFactory,
		f.loadEventLogger(),
		f.logger,
	), nil
}

func (f *factory) loadCMDRunner() boshsys.CmdRunner {
	if f.runner != nil {
		return f.runner
	}
	f.runner = boshsys.NewExecCmdRunner(f.logger)
	return f.runner
}

func (f *factory) loadCompressor() boshcmd.Compressor {
	if f.compressor != nil {
		return f.compressor
	}
	f.compressor = boshcmd.NewTarballCompressor(f.loadCMDRunner(), f.fs)
	return f.compressor
}

func (f *factory) loadStemcellRepo() bmconfig.StemcellRepo {
	if f.stemcellRepo != nil {
		return f.stemcellRepo
	}
	f.stemcellRepo = bmconfig.NewStemcellRepo(f.deploymentConfigService, f.uuidGenerator)
	return f.stemcellRepo
}

func (f *factory) loadVMRepo() bmconfig.VMRepo {
	if f.vmRepo != nil {
		return f.vmRepo
	}
	f.vmRepo = bmconfig.NewVMRepo(f.deploymentConfigService)
	return f.vmRepo
}

func (f *factory) loadDiskRepo() bmconfig.DiskRepo {
	if f.diskRepo != nil {
		return f.diskRepo
	}
	f.diskRepo = bmconfig.NewDiskRepo(f.deploymentConfigService, f.uuidGenerator)
	return f.diskRepo
}

func (f *factory) loadRegistryServerFactory() bmregistry.ServerFactory {
	if f.registryServerFactory != nil {
		return f.registryServerFactory
	}

	f.registryServerFactory = bmregistry.NewServerFactory(f.logger)
	return f.registryServerFactory
}

func (f *factory) loadSSHTunnelFactory() bmsshtunnel.Factory {
	if f.sshTunnelFactory != nil {
		return f.sshTunnelFactory
	}

	f.sshTunnelFactory = bmsshtunnel.NewFactory(f.logger)
	return f.sshTunnelFactory
}

func (f *factory) loadDiskDeployer() bminstance.DiskDeployer {
	if f.diskDeployer != nil {
		return f.diskDeployer
	}

	f.diskDeployer = bminstance.NewDiskDeployer(f.loadDiskManagerFactory(), f.loadDiskRepo(), f.logger)
	return f.diskDeployer
}

func (f *factory) loadDiskManagerFactory() bmdisk.ManagerFactory {
	if f.diskManagerFactory != nil {
		return f.diskManagerFactory
	}

	f.diskManagerFactory = bmdisk.NewManagerFactory(f.loadDiskRepo(), f.logger)
	return f.diskManagerFactory
}

func (f *factory) loadInstanceManagerFactory() bminstance.ManagerFactory {
	if f.instanceManagerFactory != nil {
		return f.instanceManagerFactory
	}

	f.instanceManagerFactory = bminstance.NewManagerFactory(
		f.loadRegistryServerFactory(),
		f.loadSSHTunnelFactory(),
		f.loadDiskDeployer(),
		f.logger,
	)
	return f.instanceManagerFactory
}

func (f *factory) loadVMManagerFactory() bmvm.ManagerFactory {
	if f.vmManagerFactory != nil {
		return f.vmManagerFactory
	}

	erbRenderer := bmerbrenderer.NewERBRenderer(f.fs, f.loadCMDRunner(), f.logger)
	jobRenderer := bmtempcomp.NewJobRenderer(erbRenderer, f.fs, f.logger)

	agentClientFactory := bmagentclient.NewAgentClientFactory(f.deploymentWorkspace.DeploymentUUID(), 1*time.Second, f.logger)
	blobstoreFactory := bmblobstore.NewBlobstoreFactory(f.fs, f.logger)
	sha1Calculator := bmcrypto.NewSha1Calculator(f.fs)
	applySpecFactory := bmas.NewFactory()

	templatesSpecGenerator := bmas.NewTemplatesSpecGenerator(
		blobstoreFactory,
		f.loadCompressor(),
		jobRenderer,
		f.uuidGenerator,
		sha1Calculator,
		f.fs,
		f.logger,
	)

	f.vmManagerFactory = bmvm.NewManagerFactory(
		f.loadVMRepo(),
		f.loadStemcellRepo(),
		agentClientFactory,
		applySpecFactory,
		templatesSpecGenerator,
		f.fs,
		f.logger,
	)
	return f.vmManagerFactory
}

func (f *factory) loadStemcellManagerFactory() bmstemcell.ManagerFactory {
	if f.stemcellManagerFactory != nil {
		return f.stemcellManagerFactory
	}

	f.stemcellManagerFactory = bmstemcell.NewManagerFactory(f.loadStemcellRepo(), f.loadEventLogger())
	return f.stemcellManagerFactory
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

func (f *factory) loadEventLogger() bmeventlog.EventLogger {
	if f.eventLogger != nil {
		return f.eventLogger
	}

	eventFilters := []bmeventlog.EventFilter{bmeventlog.NewTimeFilter(f.loadTimeService())}
	f.eventLogger = bmeventlog.NewEventLoggerWithFilters(f.ui, eventFilters)
	return f.eventLogger
}

func (f *factory) loadTimeService() boshtime.Service {
	if f.timeService != nil {
		return f.timeService
	}

	f.timeService = boshtime.NewConcreteService()
	return f.timeService
}

func (f *factory) loadCPIInstaller() bmcpi.Installer {
	if f.cpiInstaller != nil {
		return f.cpiInstaller
	}

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

	da := bmcomp.NewDependencyAnalysis()
	releasePackagesCompiler := bmcomp.NewReleasePackagesCompiler(
		da,
		packageCompiler,
		f.loadEventLogger(),
		f.loadTimeService(),
	)

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
		f.loadEventLogger(),
		f.loadTimeService(),
	)
	cloudFactory := bmcloud.NewFactory(f.fs, runner, f.deploymentWorkspace, f.logger)
	f.cpiInstaller = bmcpi.NewInstaller(
		f.ui,
		f.fs,
		extractor,
		releaseValidator,
		releaseCompiler,
		jobInstaller,
		cloudFactory,
		f.logger,
	)
	return f.cpiInstaller
}
