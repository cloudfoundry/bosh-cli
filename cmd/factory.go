package cmd

import (
	"errors"
	"path/filepath"
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	biblobstore "github.com/cloudfoundry/bosh-init/blobstore"
	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bicrypto "github.com/cloudfoundry/bosh-init/crypto"
	bidepl "github.com/cloudfoundry/bosh-init/deployment"
	bihttpagent "github.com/cloudfoundry/bosh-init/deployment/agentclient/http"
	bidisk "github.com/cloudfoundry/bosh-init/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-init/deployment/instance"
	biinstancestate "github.com/cloudfoundry/bosh-init/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bideplrel "github.com/cloudfoundry/bosh-init/deployment/release"
	bisshtunnel "github.com/cloudfoundry/bosh-init/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-init/deployment/vm"
	biindex "github.com/cloudfoundry/bosh-init/index"
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biregistry "github.com/cloudfoundry/bosh-init/registry"
	birel "github.com/cloudfoundry/bosh-init/release"
	birelset "github.com/cloudfoundry/bosh-init/release/set"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bistatepkg "github.com/cloudfoundry/bosh-init/state/pkg"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	bitemplate "github.com/cloudfoundry/bosh-init/templatescompiler"
	bitemplateerb "github.com/cloudfoundry/bosh-init/templatescompiler/erbrenderer"
	biui "github.com/cloudfoundry/bosh-init/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands                       map[string](func() (Cmd, error))
	userConfig                     biconfig.UserConfig
	userConfigService              biconfig.UserConfigService
	legacyDeploymentConfigMigrator biconfig.LegacyDeploymentConfigMigrator
	deploymentConfigService        biconfig.DeploymentConfigService
	fs                             boshsys.FileSystem
	ui                             biui.UI
	timeService                    boshtime.Service
	logger                         boshlog.Logger
	uuidGenerator                  boshuuid.Generator
	workspaceRootPath              string
	runner                         boshsys.CmdRunner
	compressor                     boshcmd.Compressor
	agentClientFactory             bihttpagent.AgentClientFactory
	vmManagerFactory               bivm.ManagerFactory
	vmRepo                         biconfig.VMRepo
	stemcellRepo                   biconfig.StemcellRepo
	diskRepo                       biconfig.DiskRepo
	registryServerManager          biregistry.ServerManager
	sshTunnelFactory               bisshtunnel.Factory
	diskDeployer                   bivm.DiskDeployer
	diskManagerFactory             bidisk.ManagerFactory
	instanceFactory                biinstance.Factory
	instanceManagerFactory         biinstance.ManagerFactory
	stemcellManagerFactory         bistemcell.ManagerFactory
	deploymentManagerFactory       bidepl.ManagerFactory
	deploymentFactory              bidepl.Factory
	deployer                       bidepl.Deployer
	blobstoreFactory               biblobstore.Factory
	eventLogger                    biui.Stage
	installerFactory               biinstall.InstallerFactory
	releaseExtractor               birel.Extractor
	releaseManager                 birel.Manager
	releaseResolver                birelset.Resolver
	releaseSetParser               birelsetmanifest.Parser
	releaseJobResolver             bideplrel.JobResolver
	installationParser             biinstallmanifest.Parser
	deploymentParser               bideplmanifest.Parser
	releaseSetValidator            birelsetmanifest.Validator
	installationValidator          biinstallmanifest.Validator
	deploymentValidator            bideplmanifest.Validator
	cloudFactory                   bicloud.Factory
	stateBuilderFactory            biinstancestate.BuilderFactory
	compiledPackageRepo            bistatepkg.CompiledPackageRepo
}

func NewFactory(
	userConfig biconfig.UserConfig,
	userConfigService biconfig.UserConfigService,
	fs boshsys.FileSystem,
	ui biui.UI,
	timeService boshtime.Service,
	logger boshlog.Logger,
	uuidGenerator boshuuid.Generator,
	workspaceRootPath string,
) Factory {
	f := &factory{
		userConfig:        userConfig,
		userConfigService: userConfigService,
		fs:                fs,
		ui:                ui,
		timeService:       timeService,
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
	stemcellReader := bistemcell.NewReader(f.loadCompressor(), f.fs)
	stemcellExtractor := bistemcell.NewExtractor(stemcellReader, f.fs)

	deploymentRepo := biconfig.NewDeploymentRepo(f.loadDeploymentConfigService())
	releaseRepo := biconfig.NewReleaseRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	sha1Calculator := bicrypto.NewSha1Calculator(f.fs)
	deploymentRecord := bidepl.NewRecord(deploymentRepo, releaseRepo, f.loadStemcellRepo(), sha1Calculator)

	return NewDeployCmd(
		f.ui,
		f.userConfig,
		f.fs,
		f.loadReleaseSetParser(),
		f.loadInstallationParser(),
		f.loadDeploymentParser(),
		f.loadLegacyDeploymentConfigMigrator(),
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

func (f *factory) loadStemcellRepo() biconfig.StemcellRepo {
	if f.stemcellRepo != nil {
		return f.stemcellRepo
	}
	f.stemcellRepo = biconfig.NewStemcellRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	return f.stemcellRepo
}

func (f *factory) loadVMRepo() biconfig.VMRepo {
	if f.vmRepo != nil {
		return f.vmRepo
	}
	f.vmRepo = biconfig.NewVMRepo(f.loadDeploymentConfigService())
	return f.vmRepo
}

func (f *factory) loadDiskRepo() biconfig.DiskRepo {
	if f.diskRepo != nil {
		return f.diskRepo
	}
	f.diskRepo = biconfig.NewDiskRepo(f.loadDeploymentConfigService(), f.uuidGenerator)
	return f.diskRepo
}

func (f *factory) loadCompiledPackageRepo() bistatepkg.CompiledPackageRepo {
	if f.compiledPackageRepo != nil {
		return f.compiledPackageRepo
	}

	index := biindex.NewInMemoryIndex()
	f.compiledPackageRepo = bistatepkg.NewCompiledPackageRepo(index)
	return f.compiledPackageRepo
}

func (f *factory) loadRegistryServerManager() biregistry.ServerManager {
	if f.registryServerManager != nil {
		return f.registryServerManager
	}

	f.registryServerManager = biregistry.NewServerManager(f.logger)
	return f.registryServerManager
}

func (f *factory) loadSSHTunnelFactory() bisshtunnel.Factory {
	if f.sshTunnelFactory != nil {
		return f.sshTunnelFactory
	}

	f.sshTunnelFactory = bisshtunnel.NewFactory(f.logger)
	return f.sshTunnelFactory
}

func (f *factory) loadDiskDeployer() bivm.DiskDeployer {
	if f.diskDeployer != nil {
		return f.diskDeployer
	}

	f.diskDeployer = bivm.NewDiskDeployer(f.loadDiskManagerFactory(), f.loadDiskRepo(), f.logger)
	return f.diskDeployer
}

func (f *factory) loadDiskManagerFactory() bidisk.ManagerFactory {
	if f.diskManagerFactory != nil {
		return f.diskManagerFactory
	}

	f.diskManagerFactory = bidisk.NewManagerFactory(f.loadDiskRepo(), f.logger)
	return f.diskManagerFactory
}

func (f *factory) loadInstanceManagerFactory() biinstance.ManagerFactory {
	if f.instanceManagerFactory != nil {
		return f.instanceManagerFactory
	}

	f.instanceManagerFactory = biinstance.NewManagerFactory(
		f.loadSSHTunnelFactory(),
		f.loadInstanceFactory(),
		f.logger,
	)
	return f.instanceManagerFactory
}

func (f *factory) loadInstanceFactory() biinstance.Factory {
	if f.instanceFactory != nil {
		return f.instanceFactory
	}

	f.instanceFactory = biinstance.NewFactory(
		f.loadBuilderFactory(),
	)
	return f.instanceFactory
}

func (f *factory) loadReleaseJobResolver() bideplrel.JobResolver {
	if f.releaseJobResolver != nil {
		return f.releaseJobResolver
	}

	releaseSetResolver := birelset.NewResolver(f.loadReleaseManager(), f.logger)
	f.releaseJobResolver = bideplrel.NewJobResolver(releaseSetResolver)
	return f.releaseJobResolver
}

func (f *factory) loadBuilderFactory() biinstancestate.BuilderFactory {
	if f.stateBuilderFactory != nil {
		return f.stateBuilderFactory
	}

	erbRenderer := bitemplateerb.NewERBRenderer(f.fs, f.loadCMDRunner(), f.logger)
	jobRenderer := bitemplate.NewJobRenderer(erbRenderer, f.fs, f.logger)
	jobListRenderer := bitemplate.NewJobListRenderer(jobRenderer, f.logger)

	sha1Calculator := bicrypto.NewSha1Calculator(f.fs)

	renderedJobListCompressor := bitemplate.NewRenderedJobListCompressor(
		f.fs,
		f.loadCompressor(),
		sha1Calculator,
		f.logger,
	)

	f.stateBuilderFactory = biinstancestate.NewBuilderFactory(
		f.loadCompiledPackageRepo(),
		f.loadReleaseJobResolver(),
		jobListRenderer,
		renderedJobListCompressor,
		f.logger,
	)
	return f.stateBuilderFactory
}

func (f *factory) loadDeploymentManagerFactory() bidepl.ManagerFactory {
	if f.deploymentManagerFactory != nil {
		return f.deploymentManagerFactory
	}

	f.deploymentManagerFactory = bidepl.NewManagerFactory(
		f.loadVMManagerFactory(),
		f.loadInstanceManagerFactory(),
		f.loadDiskManagerFactory(),
		f.loadStemcellManagerFactory(),
		f.loadDeploymentFactory(),
	)
	return f.deploymentManagerFactory
}

func (f *factory) loadDeploymentFactory() bidepl.Factory {
	if f.deploymentFactory != nil {
		return f.deploymentFactory
	}

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond

	f.deploymentFactory = bidepl.NewFactory(
		pingTimeout,
		pingDelay,
	)
	return f.deploymentFactory
}

func (f *factory) loadAgentClientFactory() bihttpagent.AgentClientFactory {
	if f.agentClientFactory != nil {
		return f.agentClientFactory
	}

	f.agentClientFactory = bihttpagent.NewAgentClientFactory(1*time.Second, f.logger)
	return f.agentClientFactory
}

func (f *factory) loadVMManagerFactory() bivm.ManagerFactory {
	if f.vmManagerFactory != nil {
		return f.vmManagerFactory
	}

	f.vmManagerFactory = bivm.NewManagerFactory(
		f.loadVMRepo(),
		f.loadStemcellRepo(),
		f.loadDiskDeployer(),
		f.uuidGenerator,
		f.fs,
		f.logger,
	)
	return f.vmManagerFactory
}

func (f *factory) loadStemcellManagerFactory() bistemcell.ManagerFactory {
	if f.stemcellManagerFactory != nil {
		return f.stemcellManagerFactory
	}

	f.stemcellManagerFactory = bistemcell.NewManagerFactory(f.loadStemcellRepo())
	return f.stemcellManagerFactory
}

func (f *factory) loadLegacyDeploymentConfigMigrator() biconfig.LegacyDeploymentConfigMigrator {
	if f.legacyDeploymentConfigMigrator != nil {
		return f.legacyDeploymentConfigMigrator
	}

	if !f.userConfig.IsDeploymentSet() {
		// no deployment set.
		// each cmd should handle validation before using deploymentConfigService or deploymentWorkspace
		return nil
	}

	f.legacyDeploymentConfigMigrator = biconfig.NewLegacyDeploymentConfigMigrator(
		f.userConfig.LegacyDeploymentConfigPath(),
		f.loadDeploymentConfigService(),
		f.fs,
		f.uuidGenerator,
		f.logger,
	)
	return f.legacyDeploymentConfigMigrator
}

func (f *factory) loadDeploymentConfigService() biconfig.DeploymentConfigService {
	if f.deploymentConfigService != nil {
		return f.deploymentConfigService
	}

	if !f.userConfig.IsDeploymentSet() {
		// no deployment set.
		// each cmd should handle validation before using deploymentConfigService or deploymentWorkspace
		return nil
	}

	f.deploymentConfigService = biconfig.NewFileSystemDeploymentConfigService(
		f.userConfig.DeploymentConfigPath(),
		f.fs,
		f.uuidGenerator,
		f.logger,
	)
	return f.deploymentConfigService
}

func (f *factory) loadDeployer() bidepl.Deployer {
	if f.deployer != nil {
		return f.deployer
	}

	f.deployer = bidepl.NewDeployer(
		f.loadVMManagerFactory(),
		f.loadInstanceManagerFactory(),
		f.loadDeploymentFactory(),
		f.logger,
	)
	return f.deployer
}

func (f *factory) loadBlobstoreFactory() biblobstore.Factory {
	if f.blobstoreFactory != nil {
		return f.blobstoreFactory
	}

	f.blobstoreFactory = biblobstore.NewBlobstoreFactory(f.uuidGenerator, f.fs, f.logger)
	return f.blobstoreFactory
}

func (f *factory) loadReleaseExtractor() birel.Extractor {
	if f.releaseExtractor != nil {
		return f.releaseExtractor
	}

	releaseValidator := birel.NewValidator(f.fs)
	f.releaseExtractor = birel.NewExtractor(f.fs, f.loadCompressor(), releaseValidator, f.logger)
	return f.releaseExtractor
}

func (f *factory) loadReleaseManager() birel.Manager {
	if f.releaseManager != nil {
		return f.releaseManager
	}

	f.releaseManager = birel.NewManager(f.logger)
	return f.releaseManager
}

func (f *factory) loadReleaseResolver() birelset.Resolver {
	if f.releaseResolver != nil {
		return f.releaseResolver
	}

	f.releaseResolver = birelset.NewResolver(f.loadReleaseManager(), f.logger)
	return f.releaseResolver
}

func (f *factory) loadReleaseSetParser() birelsetmanifest.Parser {
	if f.releaseSetParser != nil {
		return f.releaseSetParser
	}

	f.releaseSetParser = birelsetmanifest.NewParser(f.fs, f.logger)
	return f.releaseSetParser
}

func (f *factory) loadInstallationParser() biinstallmanifest.Parser {
	if f.installationParser != nil {
		return f.installationParser
	}

	f.installationParser = biinstallmanifest.NewParser(f.fs, f.logger)
	return f.installationParser
}

func (f *factory) loadDeploymentParser() bideplmanifest.Parser {
	if f.deploymentParser != nil {
		return f.deploymentParser
	}

	f.deploymentParser = bideplmanifest.NewParser(f.fs, f.logger)
	return f.deploymentParser
}

func (f *factory) loadInstallationValidator() biinstallmanifest.Validator {
	if f.installationValidator != nil {
		return f.installationValidator
	}

	f.installationValidator = biinstallmanifest.NewValidator(f.logger, f.loadReleaseResolver())
	return f.installationValidator
}

func (f *factory) loadDeploymentValidator() bideplmanifest.Validator {
	if f.deploymentValidator != nil {
		return f.deploymentValidator
	}

	f.deploymentValidator = bideplmanifest.NewValidator(f.logger, f.loadReleaseResolver())
	return f.deploymentValidator
}

func (f *factory) loadReleaseSetValidator() birelsetmanifest.Validator {
	if f.releaseSetValidator != nil {
		return f.releaseSetValidator
	}

	f.releaseSetValidator = birelsetmanifest.NewValidator(f.logger, f.loadReleaseResolver())
	return f.releaseSetValidator
}

func (f *factory) loadCloudFactory() bicloud.Factory {
	if f.cloudFactory != nil {
		return f.cloudFactory
	}

	f.cloudFactory = bicloud.NewFactory(f.fs, f.loadCMDRunner(), f.logger)
	return f.cloudFactory
}

func (f *factory) loadInstallerFactory() biinstall.InstallerFactory {
	if f.installerFactory != nil {
		return f.installerFactory
	}

	targetProvider := biinstall.NewTargetProvider(
		f.loadDeploymentConfigService(),
		f.uuidGenerator,
		filepath.Join(f.workspaceRootPath, "installations"),
	)

	f.installerFactory = biinstall.NewInstallerFactory(
		targetProvider,
		f.ui,
		f.fs,
		f.loadCMDRunner(),
		f.loadCompressor(),
		f.loadReleaseResolver(),
		f.loadReleaseJobResolver(),
		f.uuidGenerator,
		f.loadRegistryServerManager(),
		f.logger,
	)
	return f.installerFactory
}
