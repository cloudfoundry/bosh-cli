package cmd

import (
	"errors"
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bminstancestate "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplrel "github.com/cloudfoundry/bosh-micro-cli/deployment/release"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"
	bmtemplate "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
	bmtemplateerb "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands                 map[string](func() (Cmd, error))
	userConfig               bmconfig.UserConfig
	userConfigService        bmconfig.UserConfigService
	deploymentConfigService  bmconfig.DeploymentConfigService
	fs                       boshsys.FileSystem
	ui                       bmui.UI
	logger                   boshlog.Logger
	uuidGenerator            boshuuid.Generator
	workspaceRootPath        string
	runner                   boshsys.CmdRunner
	compressor               boshcmd.Compressor
	agentClientFactory       bmhttpagent.AgentClientFactory
	vmManagerFactory         bmvm.ManagerFactory
	vmRepo                   bmconfig.VMRepo
	stemcellRepo             bmconfig.StemcellRepo
	diskRepo                 bmconfig.DiskRepo
	registryServerManager    bmregistry.ServerManager
	sshTunnelFactory         bmsshtunnel.Factory
	diskDeployer             bmvm.DiskDeployer
	diskManagerFactory       bmdisk.ManagerFactory
	instanceFactory          bminstance.Factory
	instanceManagerFactory   bminstance.ManagerFactory
	stemcellManagerFactory   bmstemcell.ManagerFactory
	deploymentManagerFactory bmdepl.ManagerFactory
	deploymentFactory        bmdepl.Factory
	deployer                 bmdepl.Deployer
	blobstoreFactory         bmblobstore.Factory
	eventLogger              bmeventlog.EventLogger
	timeService              boshtime.Service
	installerFactory         bminstall.InstallerFactory
	releaseExtractor         bmrel.Extractor
	releaseManager           bmrel.Manager
	releaseResolver          bmrelset.Resolver
	releaseSetParser         bmrelsetmanifest.Parser
	installationParser       bminstallmanifest.Parser
	deploymentParser         bmdeplmanifest.Parser
	releaseSetValidator      bmrelsetmanifest.Validator
	installationValidator    bminstallmanifest.Validator
	deploymentValidator      bmdeplmanifest.Validator
	cloudFactory             bmcloud.Factory
	stateBuilderFactory      bminstancestate.BuilderFactory
}

func NewFactory(
	userConfig bmconfig.UserConfig,
	userConfigService bmconfig.UserConfigService,
	fs boshsys.FileSystem,
	ui bmui.UI,
	logger boshlog.Logger,
	uuidGenerator boshuuid.Generator,
	workspaceRootPath string,
) Factory {
	f := &factory{
		userConfig:        userConfig,
		userConfigService: userConfigService,
		fs:                fs,
		ui:                ui,
		logger:            logger,
		uuidGenerator:     uuidGenerator,
		workspaceRootPath: workspaceRootPath,
	}
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
		f.fs,
		f.uuidGenerator,
		f.logger,
	), nil
}

func (f *factory) createDeployCmd() (Cmd, error) {
	stemcellReader := bmstemcell.NewReader(f.loadCompressor(), f.fs)
	stemcellExtractor := bmstemcell.NewExtractor(stemcellReader, f.fs)

	deploymentRepo := bmconfig.NewDeploymentRepo(f.loadDeploymentConfigService())
	releaseRepo := bmconfig.NewReleaseRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	sha1Calculator := bmcrypto.NewSha1Calculator(f.fs)
	deploymentRecord := bmdepl.NewRecord(deploymentRepo, releaseRepo, f.loadStemcellRepo(), sha1Calculator)

	return NewDeployCmd(
		f.ui,
		f.userConfig,
		f.fs,
		f.loadReleaseSetParser(),
		f.loadInstallationParser(),
		f.loadDeploymentParser(),
		f.loadDeploymentConfigService(),
		f.loadReleaseSetValidator(),
		f.loadInstallationValidator(),
		f.loadDeploymentValidator(),
		f.loadInstallerFactory(),
		f.loadReleaseExtractor(),
		f.loadReleaseManager(),
		f.loadReleaseResolver(),
		f.loadCloudFactory(),
		f.loadAgentClientFactory(),
		f.loadVMManagerFactory(),
		stemcellExtractor,
		f.loadStemcellManagerFactory(),
		deploymentRecord,
		f.loadBlobstoreFactory(),
		f.loadDeployer(),
		f.loadEventLogger(),
		f.logger,
	), nil
}

func (f *factory) createDeleteCmd() (Cmd, error) {
	return NewDeleteCmd(
		f.ui,
		f.userConfig,
		f.fs,
		f.loadReleaseSetParser(),
		f.loadInstallationParser(),
		f.loadDeploymentConfigService(),
		f.loadReleaseSetValidator(),
		f.loadInstallationValidator(),
		f.loadInstallerFactory(),
		f.loadReleaseExtractor(),
		f.loadReleaseManager(),
		f.loadReleaseResolver(),
		f.loadCloudFactory(),
		f.loadAgentClientFactory(),
		f.loadBlobstoreFactory(),
		f.loadDeploymentManagerFactory(),
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
	f.stemcellRepo = bmconfig.NewStemcellRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	return f.stemcellRepo
}

func (f *factory) loadVMRepo() bmconfig.VMRepo {
	if f.vmRepo != nil {
		return f.vmRepo
	}
	f.vmRepo = bmconfig.NewVMRepo(f.loadDeploymentConfigService())
	return f.vmRepo
}

func (f *factory) loadDiskRepo() bmconfig.DiskRepo {
	if f.diskRepo != nil {
		return f.diskRepo
	}
	f.diskRepo = bmconfig.NewDiskRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	return f.diskRepo
}

func (f *factory) loadRegistryServerManager() bmregistry.ServerManager {
	if f.registryServerManager != nil {
		return f.registryServerManager
	}

	f.registryServerManager = bmregistry.NewServerManager(f.logger)
	return f.registryServerManager
}

func (f *factory) loadSSHTunnelFactory() bmsshtunnel.Factory {
	if f.sshTunnelFactory != nil {
		return f.sshTunnelFactory
	}

	f.sshTunnelFactory = bmsshtunnel.NewFactory(f.logger)
	return f.sshTunnelFactory
}

func (f *factory) loadDiskDeployer() bmvm.DiskDeployer {
	if f.diskDeployer != nil {
		return f.diskDeployer
	}

	f.diskDeployer = bmvm.NewDiskDeployer(f.loadDiskManagerFactory(), f.loadDiskRepo(), f.logger)
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
		f.loadSSHTunnelFactory(),
		f.loadInstanceFactory(),
		f.logger,
	)
	return f.instanceManagerFactory
}

func (f *factory) loadInstanceFactory() bminstance.Factory {
	if f.instanceFactory != nil {
		return f.instanceFactory
	}

	f.instanceFactory = bminstance.NewFactory(
		f.loadBuilderFactory(),
	)
	return f.instanceFactory
}

func (f *factory) loadBuilderFactory() bminstancestate.BuilderFactory {
	if f.stateBuilderFactory != nil {
		return f.stateBuilderFactory
	}

	releaseSetResolver := bmrelset.NewResolver(f.loadReleaseManager(), f.logger)
	releaseJobResolver := bmdeplrel.NewJobResolver(releaseSetResolver)

	erbRenderer := bmtemplateerb.NewERBRenderer(f.fs, f.loadCMDRunner(), f.logger)
	jobRenderer := bmtemplate.NewJobRenderer(erbRenderer, f.fs, f.logger)
	jobListRenderer := bmtemplate.NewJobListRenderer(jobRenderer, f.logger)

	sha1Calculator := bmcrypto.NewSha1Calculator(f.fs)

	renderedJobListCompressor := bmtemplate.NewRenderedJobListCompressor(
		f.fs,
		f.loadCompressor(),
		sha1Calculator,
		f.logger,
	)

	f.stateBuilderFactory = bminstancestate.NewBuilderFactory(
		releaseJobResolver,
		jobListRenderer,
		renderedJobListCompressor,
		f.logger,
	)
	return f.stateBuilderFactory
}

func (f *factory) loadDeploymentManagerFactory() bmdepl.ManagerFactory {
	if f.deploymentManagerFactory != nil {
		return f.deploymentManagerFactory
	}

	f.deploymentManagerFactory = bmdepl.NewManagerFactory(
		f.loadVMManagerFactory(),
		f.loadInstanceManagerFactory(),
		f.loadDiskManagerFactory(),
		f.loadStemcellManagerFactory(),
		f.loadDeploymentFactory(),
	)
	return f.deploymentManagerFactory
}

func (f *factory) loadDeploymentFactory() bmdepl.Factory {
	if f.deploymentFactory != nil {
		return f.deploymentFactory
	}

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond

	f.deploymentFactory = bmdepl.NewFactory(
		pingTimeout,
		pingDelay,
	)
	return f.deploymentFactory
}

func (f *factory) loadAgentClientFactory() bmhttpagent.AgentClientFactory {
	if f.agentClientFactory != nil {
		return f.agentClientFactory
	}

	f.agentClientFactory = bmhttpagent.NewAgentClientFactory(1*time.Second, f.logger)
	return f.agentClientFactory
}

func (f *factory) loadVMManagerFactory() bmvm.ManagerFactory {
	if f.vmManagerFactory != nil {
		return f.vmManagerFactory
	}

	f.vmManagerFactory = bmvm.NewManagerFactory(
		f.loadVMRepo(),
		f.loadStemcellRepo(),
		f.loadDiskDeployer(),
		f.uuidGenerator,
		f.fs,
		f.logger,
	)
	return f.vmManagerFactory
}

func (f *factory) loadStemcellManagerFactory() bmstemcell.ManagerFactory {
	if f.stemcellManagerFactory != nil {
		return f.stemcellManagerFactory
	}

	f.stemcellManagerFactory = bmstemcell.NewManagerFactory(f.loadStemcellRepo())
	return f.stemcellManagerFactory
}

func (f *factory) loadDeploymentConfigService() bmconfig.DeploymentConfigService {
	if f.deploymentConfigService != nil {
		return f.deploymentConfigService
	}

	deploymentConfigPath := f.userConfig.DeploymentConfigPath()
	f.logger.Debug("cmdFactory", "DeploymentConfigPath: %s", deploymentConfigPath)
	if deploymentConfigPath == "" {
		// no deployment set.
		// each cmd should handle validation before using deploymentConfigService or deploymentWorkspace
		return nil
	}

	f.deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(
		deploymentConfigPath,
		f.fs,
		f.uuidGenerator,
		f.logger,
	)
	return f.deploymentConfigService
}

func (f *factory) loadDeployer() bmdepl.Deployer {
	if f.deployer != nil {
		return f.deployer
	}

	f.deployer = bmdepl.NewDeployer(
		f.loadVMManagerFactory(),
		f.loadInstanceManagerFactory(),
		f.loadDeploymentFactory(),
		f.loadEventLogger(),
		f.logger,
	)
	return f.deployer
}

func (f *factory) loadBlobstoreFactory() bmblobstore.Factory {
	if f.blobstoreFactory != nil {
		return f.blobstoreFactory
	}

	f.blobstoreFactory = bmblobstore.NewBlobstoreFactory(f.uuidGenerator, f.fs, f.logger)
	return f.blobstoreFactory
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

func (f *factory) loadReleaseExtractor() bmrel.Extractor {
	if f.releaseExtractor != nil {
		return f.releaseExtractor
	}

	releaseValidator := bmrel.NewValidator(f.fs)
	f.releaseExtractor = bmrel.NewExtractor(f.fs, f.loadCompressor(), releaseValidator, f.logger)
	return f.releaseExtractor
}

func (f *factory) loadReleaseManager() bmrel.Manager {
	if f.releaseManager != nil {
		return f.releaseManager
	}

	f.releaseManager = bmrel.NewManager(f.logger)
	return f.releaseManager
}

func (f *factory) loadReleaseResolver() bmrelset.Resolver {
	if f.releaseResolver != nil {
		return f.releaseResolver
	}

	f.releaseResolver = bmrelset.NewResolver(f.loadReleaseManager(), f.logger)
	return f.releaseResolver
}

func (f *factory) loadReleaseSetParser() bmrelsetmanifest.Parser {
	if f.releaseSetParser != nil {
		return f.releaseSetParser
	}

	f.releaseSetParser = bmrelsetmanifest.NewParser(f.fs, f.logger)
	return f.releaseSetParser
}

func (f *factory) loadInstallationParser() bminstallmanifest.Parser {
	if f.installationParser != nil {
		return f.installationParser
	}

	f.installationParser = bminstallmanifest.NewParser(f.fs, f.logger)
	return f.installationParser
}

func (f *factory) loadDeploymentParser() bmdeplmanifest.Parser {
	if f.deploymentParser != nil {
		return f.deploymentParser
	}

	f.deploymentParser = bmdeplmanifest.NewParser(f.fs, f.logger)
	return f.deploymentParser
}

func (f *factory) loadInstallationValidator() bminstallmanifest.Validator {
	if f.installationValidator != nil {
		return f.installationValidator
	}

	f.installationValidator = bminstallmanifest.NewValidator(f.logger, f.loadReleaseResolver())
	return f.installationValidator
}

func (f *factory) loadDeploymentValidator() bmdeplmanifest.Validator {
	if f.deploymentValidator != nil {
		return f.deploymentValidator
	}

	f.deploymentValidator = bmdeplmanifest.NewValidator(f.logger, f.loadReleaseResolver())
	return f.deploymentValidator
}

func (f *factory) loadReleaseSetValidator() bmrelsetmanifest.Validator {
	if f.releaseSetValidator != nil {
		return f.releaseSetValidator
	}

	f.releaseSetValidator = bmrelsetmanifest.NewValidator(f.logger, f.loadReleaseResolver())
	return f.releaseSetValidator
}

func (f *factory) loadCloudFactory() bmcloud.Factory {
	if f.cloudFactory != nil {
		return f.cloudFactory
	}

	f.cloudFactory = bmcloud.NewFactory(f.fs, f.loadCMDRunner(), f.logger)
	return f.cloudFactory
}

func (f *factory) loadInstallerFactory() bminstall.InstallerFactory {
	if f.installerFactory != nil {
		return f.installerFactory
	}

	f.installerFactory = bminstall.NewInstallerFactory(
		f.ui,
		f.fs,
		f.loadCMDRunner(),
		f.loadCompressor(),
		f.loadDeploymentConfigService(),
		f.loadReleaseResolver(),
		f.workspaceRootPath,
		f.uuidGenerator,
		f.loadTimeService(),
		f.loadRegistryServerManager(),
		f.loadEventLogger(),
		f.logger,
	)
	return f.installerFactory
}
