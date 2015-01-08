package cmd

import (
	"errors"
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmhttpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/deployment/blobstore"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/validator"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
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
	deploymentConfigService bmconfig.DeploymentConfigService
	fs                      boshsys.FileSystem
	ui                      bmui.UI
	logger                  boshlog.Logger
	uuidGenerator           boshuuid.Generator
	workspaceRootPath       string
	runner                  boshsys.CmdRunner
	compressor              boshcmd.Compressor
	agentClientFactory      bmhttpagent.AgentClientFactory
	vmManagerFactory        bmvm.ManagerFactory
	vmRepo                  bmconfig.VMRepo
	stemcellRepo            bmconfig.StemcellRepo
	diskRepo                bmconfig.DiskRepo
	registryServerManager   bmregistry.ServerManager
	sshTunnelFactory        bmsshtunnel.Factory
	diskDeployer            bmvm.DiskDeployer
	diskManagerFactory      bmdisk.ManagerFactory
	instanceManagerFactory  bminstance.ManagerFactory
	stemcellManagerFactory  bmstemcell.ManagerFactory
	deploymentFactory       bmdepl.Factory
	eventLogger             bmeventlog.EventLogger
	timeService             boshtime.Service
	installerFactory        bminstall.InstallerFactory
	releaseExtractor        bmrel.Extractor
	releaseManager          bmrel.Manager
	installationParser      bminstallmanifest.Parser
	deploymentParser        bmdeplmanifest.Parser
	cloudFactory            bmcloud.Factory
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
	boshDeploymentValidator := bmdeplval.NewBoshDeploymentValidator()

	stemcellReader := bmstemcell.NewReader(f.loadCompressor(), f.fs)
	stemcellExtractor := bmstemcell.NewExtractor(stemcellReader, f.fs)

	deploymentRepo := bmconfig.NewDeploymentRepo(f.loadDeploymentConfigService())
	releaseRepo := bmconfig.NewReleaseRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	sha1Calculator := bmcrypto.NewSha1Calculator(f.fs)
	deploymentRecord := bmdepl.NewDeploymentRecord(deploymentRepo, releaseRepo, f.loadStemcellRepo(), sha1Calculator)

	return NewDeployCmd(
		f.ui,
		f.userConfig,
		f.fs,
		f.loadInstallationParser(),
		f.loadDeploymentParser(),
		f.loadDeploymentConfigService(),
		boshDeploymentValidator,
		f.loadInstallerFactory(),
		f.loadReleaseExtractor(),
		f.loadReleaseManager(),
		f.loadCloudFactory(),
		f.loadAgentClientFactory(),
		f.loadVMManagerFactory(),
		stemcellExtractor,
		deploymentRecord,
		f.loadDeploymentFactory(),
		f.loadEventLogger(),
		f.logger,
	), nil
}

func (f *factory) createDeleteCmd() (Cmd, error) {
	return NewDeleteCmd(
		f.ui,
		f.userConfig,
		f.fs,
		f.loadInstallationParser(),
		f.loadDeploymentConfigService(),
		f.loadInstallerFactory(),
		f.loadReleaseExtractor(),
		f.loadReleaseManager(),
		f.loadCloudFactory(),
		f.loadAgentClientFactory(),
		f.loadVMManagerFactory(),
		f.loadInstanceManagerFactory(),
		f.loadDiskManagerFactory(),
		f.loadStemcellManagerFactory(),
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
		f.logger,
	)
	return f.instanceManagerFactory
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

	erbRenderer := bmerbrenderer.NewERBRenderer(f.fs, f.loadCMDRunner(), f.logger)
	jobRenderer := bmtempcomp.NewJobRenderer(erbRenderer, f.fs, f.logger)

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
		f.loadDiskDeployer(),
		applySpecFactory,
		templatesSpecGenerator,
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

	f.stemcellManagerFactory = bmstemcell.NewManagerFactory(f.loadStemcellRepo(), f.loadEventLogger())
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

func (f *factory) loadDeploymentFactory() bmdepl.Factory {
	if f.deploymentFactory != nil {
		return f.deploymentFactory
	}

	deployer := bmdepl.NewDeployer(
		f.loadStemcellManagerFactory(),
		f.loadVMManagerFactory(),
		f.loadSSHTunnelFactory(),
		f.loadEventLogger(),
		f.logger,
	)
	f.deploymentFactory = bmdepl.NewFactory(deployer)
	return f.deploymentFactory
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

	boshReleaseValidator := bmrelvalidation.NewBoshValidator(f.fs)
	f.releaseExtractor = bmrel.NewExtractor(f.fs, f.loadCompressor(), boshReleaseValidator, f.logger)
	return f.releaseExtractor
}

func (f *factory) loadReleaseManager() bmrel.Manager {
	if f.releaseManager != nil {
		return f.releaseManager
	}

	f.releaseManager = bmrel.NewManager(f.logger)
	return f.releaseManager
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
		f.loadReleaseManager(),
		f.workspaceRootPath,
		f.uuidGenerator,
		f.loadTimeService(),
		f.loadRegistryServerManager(),
		f.loadEventLogger(),
		f.logger,
	)
	return f.installerFactory
}
