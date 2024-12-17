package cmd

import (
	"os"
	"path/filepath"
	"time"

	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	"github.com/cppforlife/go-patch/patch"

	biblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-cli/v7/deployment"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	biinstancestate "github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bideplrel "github.com/cloudfoundry/bosh-cli/v7/deployment/release"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel"
	bidepltpl "github.com/cloudfoundry/bosh-cli/v7/deployment/template"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biindex "github.com/cloudfoundry/bosh-cli/v7/index"
	boshinst "github.com/cloudfoundry/bosh-cli/v7/installation"
	boshinstmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	bitarball "github.com/cloudfoundry/bosh-cli/v7/installation/tarball"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	bistatepkg "github.com/cloudfoundry/bosh-cli/v7/state/pkg"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	bitemplate "github.com/cloudfoundry/bosh-cli/v7/templatescompiler"
	bitemplateerb "github.com/cloudfoundry/bosh-cli/v7/templatescompiler/erbrenderer"
)

type envFactory struct {
	deps         BasicDeps
	manifestPath string
	manifestVars boshtpl.Variables
	manifestOp   patch.Op

	deploymentStateService     biconfig.DeploymentStateService
	installationManifestParser ReleaseSetAndInstallationManifestParser

	releaseManager  boshinst.ReleaseManager
	releaseFetcher  boshinst.ReleaseFetcher
	stemcellFetcher bistemcell.Fetcher

	cpiInstaller   bicpirel.CpiInstaller
	targetProvider boshinst.TargetProvider
	cloudFactory   bicloud.Factory

	diskManagerFactory     bidisk.ManagerFactory
	vmManagerFactory       bivm.ManagerFactory
	stemcellManagerFactory bistemcell.ManagerFactory

	instanceManagerFactory biinstance.ManagerFactory

	agentClientFactory bihttpagent.AgentClientFactory
	blobstoreFactory   biblobstore.Factory
	deploymentFactory  bidepl.Factory
	deploymentRecord   bidepl.Record
}

func NewEnvFactory(
	deps BasicDeps,
	manifestPath string,
	statePath string,
	manifestVars boshtpl.Variables,
	manifestOp patch.Op,
	recreatePersistentDisks bool,
	packageDir string,
) *envFactory {
	f := envFactory{
		deps:         deps,
		manifestPath: manifestPath,
		manifestVars: manifestVars,
		manifestOp:   manifestOp,
	}

	f.releaseManager = boshinst.NewReleaseManager(deps.Logger)
	releaseJobResolver := bideplrel.NewJobResolver(f.releaseManager)

	// todo expand path?
	workspaceRootPath := filepath.Join(os.Getenv("HOME"), ".bosh")

	{
		tarballCacheBasePath := filepath.Join(workspaceRootPath, "downloads")
		tarballCache := bitarball.NewCache(tarballCacheBasePath, deps.FS, deps.Logger)
		httpClient := httpclient.NewHTTPClient(httpclient.CreateExternalDefaultClient(nil), deps.Logger)
		tarballProvider := bitarball.NewProvider(
			tarballCache, deps.FS, httpClient, 3, 500*time.Millisecond, deps.Logger)

		releaseProvider := boshrel.NewProvider(
			deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)

		f.releaseFetcher = boshinst.NewReleaseFetcher(
			tarballProvider,
			releaseProvider.NewExtractingArchiveReader(),
			f.releaseManager,
		)

		stemcellReader := bistemcell.NewReader(deps.Compressor, deps.FS)
		stemcellExtractor := bistemcell.NewExtractor(stemcellReader, deps.FS)

		f.stemcellFetcher = bistemcell.Fetcher{
			TarballProvider:   tarballProvider,
			StemcellExtractor: stemcellExtractor,
		}
	}

	f.deploymentStateService = biconfig.NewFileSystemDeploymentStateService(
		deps.FS, deps.UUIDGen, deps.Logger, biconfig.DeploymentStatePath(manifestPath, statePath))

	{
		installerFactory := boshinst.NewInstallerFactory(
			deps.UI, deps.CmdRunner, deps.Compressor, releaseJobResolver,
			deps.UUIDGen, deps.Logger, deps.FS, deps.DigestCreationAlgorithms)

		f.cpiInstaller = bicpirel.CpiInstaller{
			ReleaseManager:   f.releaseManager,
			InstallerFactory: installerFactory,
		}
	}

	f.targetProvider = boshinst.NewTargetProvider(
		f.deploymentStateService, deps.UUIDGen, filepath.Join(workspaceRootPath, "installations"), packageDir)

	{
		diskRepo := biconfig.NewDiskRepo(f.deploymentStateService, deps.UUIDGen)
		stemcellRepo := biconfig.NewStemcellRepo(f.deploymentStateService, deps.UUIDGen)
		vmRepo := biconfig.NewVMRepo(f.deploymentStateService)

		f.diskManagerFactory = bidisk.NewManagerFactory(diskRepo, deps.Logger)
		diskDeployer := bivm.NewDiskDeployer(f.diskManagerFactory, diskRepo, deps.Logger, recreatePersistentDisks)

		f.stemcellManagerFactory = bistemcell.NewManagerFactory(stemcellRepo)
		f.vmManagerFactory = bivm.NewManagerFactory(
			vmRepo, stemcellRepo, diskDeployer, deps.UUIDGen, deps.FS, deps.Logger)

		deploymentRepo := biconfig.NewDeploymentRepo(f.deploymentStateService)
		releaseRepo := biconfig.NewReleaseRepo(f.deploymentStateService, deps.UUIDGen)
		f.deploymentRecord = bidepl.NewRecord(deploymentRepo, releaseRepo, stemcellRepo)
	}

	{
		f.blobstoreFactory = biblobstore.NewBlobstoreFactory(deps.UUIDGen, deps.FS, deps.Logger)
		f.deploymentFactory = bidepl.NewFactory(10*time.Second, 500*time.Millisecond)
		f.agentClientFactory = bihttpagent.NewAgentClientFactory(1*time.Second, deps.Logger)
		f.cloudFactory = bicloud.NewFactory(deps.FS, deps.CmdRunner, deps.Logger)
	}

	{
		erbRenderer := bitemplateerb.NewERBRenderer(deps.FS, deps.CmdRunner, deps.Logger)
		jobRenderer := bitemplate.NewJobRenderer(erbRenderer, deps.FS, deps.UUIDGen, deps.Logger)

		builderFactory := biinstancestate.NewBuilderFactory(
			bistatepkg.NewCompiledPackageRepo(biindex.NewInMemoryIndex()),
			releaseJobResolver,
			bitemplate.NewJobListRenderer(jobRenderer, deps.Logger),
			bitemplate.NewRenderedJobListCompressor(deps.FS, deps.Compressor, deps.DigestCalculator, deps.Logger),
			deps.Logger,
		)

		sshTunnelFactory := bisshtunnel.NewFactory(deps.Logger)
		instanceFactory := biinstance.NewFactory(builderFactory)

		f.instanceManagerFactory = biinstance.NewManagerFactory(
			sshTunnelFactory, instanceFactory, deps.Logger)
	}

	{
		releaseSetValidator := birelsetmanifest.NewValidator(deps.Logger)
		releaseSetParser := birelsetmanifest.NewParser(deps.FS, deps.Logger, releaseSetValidator)

		installValidator := boshinstmanifest.NewValidator(deps.Logger)
		installParser := boshinstmanifest.NewParser(deps.FS, deps.UUIDGen, deps.Logger, installValidator)

		f.installationManifestParser = ReleaseSetAndInstallationManifestParser{
			ReleaseSetParser:   releaseSetParser,
			InstallationParser: installParser,
		}
	}

	return &f
}

func (f *envFactory) Preparer() DeploymentPreparer {
	return NewDeploymentPreparer(
		f.deps.UI,
		f.deps.Logger,
		"DeploymentPreparer",
		f.deploymentStateService,
		biconfig.NewLegacyDeploymentStateMigrator(
			f.deploymentStateService,
			f.deps.FS,
			f.deps.UUIDGen,
			f.deps.Logger,
		),
		f.releaseManager,
		f.deploymentRecord,
		f.cloudFactory,
		f.stemcellManagerFactory,
		f.agentClientFactory,
		f.vmManagerFactory,
		f.blobstoreFactory,
		bidepl.NewDeployer(
			f.vmManagerFactory,
			f.instanceManagerFactory,
			f.deploymentFactory,
			f.deps.Logger,
		),
		f.manifestPath,
		f.manifestVars,
		f.manifestOp,
		f.cpiInstaller,
		f.releaseFetcher,
		f.stemcellFetcher,
		f.installationManifestParser,
		NewDeploymentManifestParser(
			bideplmanifest.NewParser(f.deps.FS, f.deps.Logger),
			bideplmanifest.NewValidator(f.deps.Logger),
			f.releaseManager,
			bidepltpl.NewDeploymentTemplateFactory(f.deps.FS),
		),
		NewTempRootConfigurator(f.deps.FS),
		f.targetProvider,
	)
}

func (f *envFactory) Deleter() DeploymentDeleter {
	return NewDeploymentDeleter(
		f.deps.UI,
		"DeploymentDeleter",
		f.deps.Logger,
		f.deploymentStateService,
		f.releaseManager,
		f.cloudFactory,
		f.agentClientFactory,
		f.blobstoreFactory,
		bidepl.NewManagerFactory(
			f.vmManagerFactory,
			f.instanceManagerFactory,
			f.diskManagerFactory,
			f.stemcellManagerFactory,
			f.deploymentFactory,
		),
		f.manifestPath,
		f.manifestVars,
		f.manifestOp,
		f.cpiInstaller,
		boshinst.NewUninstaller(f.deps.FS, f.deps.Logger),
		f.releaseFetcher,
		f.installationManifestParser,
		NewTempRootConfigurator(f.deps.FS),
		f.targetProvider,
	)
}

func (f *envFactory) StateManager() DeploymentStateManager {
	return NewDeploymentStateManager(
		f.deps.UI,
		"DeploymentStateManager",
		f.deps.Logger,
		f.deploymentStateService,
		f.agentClientFactory,
		bidepl.NewManagerFactory(
			f.vmManagerFactory,
			f.instanceManagerFactory,
			f.diskManagerFactory,
			f.stemcellManagerFactory,
			f.deploymentFactory,
		),
		f.manifestPath,
		f.manifestVars,
		f.manifestOp,
		f.installationManifestParser,
		NewDeploymentManifestParser(
			bideplmanifest.NewParser(f.deps.FS, f.deps.Logger),
			bideplmanifest.NewValidator(f.deps.Logger),
			f.releaseManager,
			bidepltpl.NewDeploymentTemplateFactory(f.deps.FS),
		),
	)
}
