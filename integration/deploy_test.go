package integration_test

import (
	. "github.com/cloudfoundry/bosh-init/cmd"

	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"
	"github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega/gbytes"

	mock_blobstore "github.com/cloudfoundry/bosh-init/blobstore/mocks"
	mock_cloud "github.com/cloudfoundry/bosh-init/cloud/mocks"
	mock_instance_state "github.com/cloudfoundry/bosh-init/deployment/instance/state/mocks"
	mock_install "github.com/cloudfoundry/bosh-init/installation/mocks"
	mock_httpagent "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/agentclient/http/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/agentclient/mocks"
	"github.com/cloudfoundry/bosh-init/internal/github.com/golang/mock/gomock"
	mock_release "github.com/cloudfoundry/bosh-init/release/mocks"

	bicloud "github.com/cloudfoundry/bosh-init/cloud"
	biconfig "github.com/cloudfoundry/bosh-init/config"
	bicpirel "github.com/cloudfoundry/bosh-init/cpi/release"
	bidepl "github.com/cloudfoundry/bosh-init/deployment"
	bidisk "github.com/cloudfoundry/bosh-init/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-init/deployment/instance"
	bideplmanifest "github.com/cloudfoundry/bosh-init/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-init/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-init/deployment/vm"
	biinstall "github.com/cloudfoundry/bosh-init/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	bitarball "github.com/cloudfoundry/bosh-init/installation/tarball"
	biagentclient "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/agentclient"
	bias "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	bihttp "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/uuid/fakes"
	biregistry "github.com/cloudfoundry/bosh-init/registry"
	birel "github.com/cloudfoundry/bosh-init/release"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	birelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
	birelsetmanifest "github.com/cloudfoundry/bosh-init/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-init/stemcell"
	biui "github.com/cloudfoundry/bosh-init/ui"

	fakebicrypto "github.com/cloudfoundry/bosh-init/crypto/fakes"
	fakebihttpclient "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/httpclient/fakes"
	fakebistemcell "github.com/cloudfoundry/bosh-init/stemcell/fakes"
	fakebiui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("bosh-init", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Deploy", func() {
		var (
			fs     *fakesys.FakeFileSystem
			logger boshlog.Logger

			registryServerManager biregistry.ServerManager
			releaseManager        birel.Manager

			mockInstaller          *mock_install.MockInstaller
			mockInstallerFactory   *mock_install.MockInstallerFactory
			mockCloudFactory       *mock_cloud.MockFactory
			mockCloud              *mock_cloud.MockCloud
			mockAgentClient        *mock_agentclient.MockAgentClient
			mockAgentClientFactory *mock_httpagent.MockAgentClientFactory
			mockReleaseExtractor   *mock_release.MockExtractor

			mockStateBuilderFactory *mock_instance_state.MockBuilderFactory
			mockStateBuilder        *mock_instance_state.MockBuilder
			mockState               *mock_instance_state.MockState

			mockBlobstoreFactory *mock_blobstore.MockFactory
			mockBlobstore        *mock_blobstore.MockBlobstore

			fakeStemcellExtractor         *fakebistemcell.FakeExtractor
			fakeUUIDGenerator             *fakeuuid.FakeGenerator
			fakeRegistryUUIDGenerator     *fakeuuid.FakeGenerator
			fakeRepoUUIDGenerator         *fakeuuid.FakeGenerator
			fakeAgentIDGenerator          *fakeuuid.FakeGenerator
			fakeSHA1Calculator            *fakebicrypto.FakeSha1Calculator
			legacyDeploymentStateMigrator biconfig.LegacyDeploymentStateMigrator
			deploymentStateService        biconfig.DeploymentStateService
			vmRepo                        biconfig.VMRepo
			diskRepo                      biconfig.DiskRepo
			stemcellRepo                  biconfig.StemcellRepo
			deploymentRepo                biconfig.DeploymentRepo
			releaseRepo                   biconfig.ReleaseRepo

			sshTunnelFactory bisshtunnel.Factory

			diskManagerFactory bidisk.ManagerFactory
			diskDeployer       bivm.DiskDeployer

			stdOut    *gbytes.Buffer
			stdErr    *gbytes.Buffer
			fakeStage *fakebiui.FakeStage

			stemcellManagerFactory bistemcell.ManagerFactory
			vmManagerFactory       bivm.ManagerFactory

			applySpec bias.ApplySpec

			directorID string

			stemcellTarballPath    = "/fake-stemcell-release.tgz"
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			deploymentStatePath    = "/deployment-dir/fake-deployment-manifest-state.json"

			stemcellImagePath       = "fake-stemcell-image-path"
			stemcellCID             = "fake-stemcell-cid"
			stemcellCloudProperties = biproperty.Map{}

			vmCloudProperties = biproperty.Map{}
			vmEnv             = biproperty.Map{}

			diskCloudProperties = biproperty.Map{}

			networkInterfaces = map[string]biproperty.Map{
				"network-1": biproperty.Map{
					"type":             "dynamic",
					"default":          []bideplmanifest.NetworkDefault{"dns", "gateway"},
					"cloud_properties": biproperty.Map{},
				},
			}

			agentRunningState = biagentclient.AgentState{JobState: "running"}
			mbusURL           = "http://fake-mbus-url"

			expectHasVM1    *gomock.Call
			expectDeleteVM1 *gomock.Call

			sshConfig *SSHConfig
		)

		var manifestTemplate = `---
name: test-deployment

releases:
- name: fake-cpi-release-name
  version: 1.1
  url: file:///fake-cpi-release.tgz

networks:
- name: network-1
  type: dynamic

resource_pools:
- name: resource-pool-1
  network: network-1
  stemcell:
    url: file:///fake-stemcell-release.tgz

jobs:
- name: fake-deployment-job-name
  instances: 1
  persistent_disk: {{ .DiskSize }}
  resource_pool: resource-pool-1
  networks:
  - name: network-1
  templates:
  - {name: fake-cpi-release-job-name, release: fake-cpi-release-name}

cloud_provider:
  template:
    name: fake-cpi-release-job-name
    release: fake-cpi-release-name
  mbus: http://fake-mbus-url
  ssh_tunnel:
    host: 127.0.0.1
    port: 22
    user: {{ .SSHTunnelUser }}
    private_key: {{ .SSHTunnelPrivateKey }}
`
		type manifestContext struct {
			DiskSize            int
			SSHTunnelUser       string
			SSHTunnelPrivateKey string
		}

		var updateManifest = func(context manifestContext) {
			buffer := bytes.NewBuffer([]byte{})
			t := template.Must(template.New("manifest").Parse(manifestTemplate))
			err := t.Execute(buffer, context)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(deploymentManifestPath, buffer.String())
			Expect(err).ToNot(HaveOccurred())
		}

		var writeDeploymentManifest = func() {
			context := manifestContext{
				DiskSize:            1024,
				SSHTunnelUser:       sshConfig.Username,
				SSHTunnelPrivateKey: sshConfig.PrivateKey,
			}
			updateManifest(context)

			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebicrypto.CalculateInput{
				deploymentManifestPath: {Sha1: "fake-deployment-sha1-1"},
			})
		}

		var writeDeploymentManifestWithLargerDisk = func() {
			context := manifestContext{
				DiskSize:            2048,
				SSHTunnelUser:       sshConfig.Username,
				SSHTunnelPrivateKey: sshConfig.PrivateKey,
			}
			updateManifest(context)

			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebicrypto.CalculateInput{
				deploymentManifestPath: {Sha1: "fake-deployment-sha1-2"},
			})
		}

		var writeCPIReleaseTarball = func() {
			err := fs.WriteFileString("/fake-cpi-release.tgz", "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowCPIToBeInstalled = func() {
			cpiPackage := birelpkg.Package{
				Name:          "fake-package-name",
				Fingerprint:   "fake-package-fingerprint-cpi",
				SHA1:          "fake-package-sha1-cpi",
				Dependencies:  []*birelpkg.Package{},
				ExtractedPath: "fake-package-extracted-path-cpi",
				ArchivePath:   "fake-package-archive-path-cpi",
			}
			cpiRelease := birel.NewRelease(
				"fake-cpi-release-name",
				"1.1",
				[]bireljob.Job{
					{
						Name: "fake-cpi-release-job-name",
						Templates: map[string]string{
							"cpi.erb": "bin/cpi",
						},
						Packages: []*birelpkg.Package{&cpiPackage},
					},
				},
				[]*birelpkg.Package{&cpiPackage},
				"fake-cpi-extracted-dir",
				fs,
			)
			mockReleaseExtractor.EXPECT().Extract("/fake-cpi-release.tgz").Do(func(_ string) {
				err := fs.MkdirAll("fake-cpi-extracted-dir", os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
			}).Return(cpiRelease, nil).AnyTimes()

			installationManifest := biinstallmanifest.Manifest{
				Name: "test-deployment",
				Template: biinstallmanifest.ReleaseJobRef{
					Name:    "fake-cpi-release-job-name",
					Release: "fake-cpi-release-name",
				},
				Registry: biinstallmanifest.Registry{
					Username: "registry",
					Password: "registry-password",
					Host:     "127.0.0.1",
					Port:     6901,
					SSHTunnel: biinstallmanifest.SSHTunnel{
						Host:       "127.0.0.1",
						Port:       22,
						User:       sshConfig.Username,
						PrivateKey: sshConfig.PrivateKey,
					},
				},
				Mbus: mbusURL,
				Properties: biproperty.Map{
					"registry": biproperty.Map{
						"username": "registry",
						"password": "registry-password",
						"host":     "127.0.0.1",
						"port":     6901,
					},
				},
			}
			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := biinstall.NewTarget(installationPath)

			installedJob := biinstall.InstalledJob{}
			installedJob.Name = "fake-cpi-release-job-name"
			installedJob.Path = filepath.Join(target.JobsPath(), "fake-cpi-release-job-name")

			installation := biinstall.NewInstallation(target, installedJob, installationManifest, registryServerManager)

			mockInstallerFactory.EXPECT().NewInstaller(target).Return(mockInstaller).AnyTimes()

			mockInstaller.EXPECT().Install(installationManifest, gomock.Any()).Do(func(_ interface{}, stage biui.Stage) {
				Expect(fakeStage.SubStages).To(ContainElement(stage))
			}).Return(installation, nil).AnyTimes()
			mockInstaller.EXPECT().Cleanup(installation).AnyTimes()
			mockCloudFactory.EXPECT().NewCloud(installation, directorID).Return(mockCloud, nil).AnyTimes()
		}

		var writeStemcellReleaseTarball = func() {
			err := fs.WriteFileString(stemcellTarballPath, "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowStemcellToBeExtracted = func() {
			stemcellManifest := bistemcell.Manifest{
				ImagePath:       "fake-stemcell-image-path",
				Name:            "fake-stemcell-name",
				Version:         "fake-stemcell-version",
				SHA1:            "fake-stemcell-sha1",
				CloudProperties: biproperty.Map{},
			}
			extractedStemcell := bistemcell.NewExtractedStemcell(
				stemcellManifest,
				"fake-stemcell-extracted-dir",
				fs,
			)
			fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, extractedStemcell, nil)
		}

		var allowApplySpecToBeCreated = func() {
			jobName := "fake-deployment-job-name"
			jobIndex := 0

			applySpec = bias.ApplySpec{
				Deployment: "test-release",
				Index:      jobIndex,
				Networks: map[string]biproperty.Map{
					"network-1": biproperty.Map{
						"cloud_properties": biproperty.Map{},
						"type":             "dynamic",
						"ip":               "",
					},
				},
				Job: bias.Job{
					Name:      jobName,
					Templates: []bias.Blob{},
				},
				Packages: map[string]bias.Blob{
					"fake-package-name": bias.Blob{
						Name:        "fake-package-name",
						Version:     "fake-package-fingerprint-cpi",
						SHA1:        "fake-compiled-package-sha1-cpi",
						BlobstoreID: "fake-compiled-package-blob-id-cpi",
					},
				},
				RenderedTemplatesArchive: bias.RenderedTemplatesArchiveSpec{},
				ConfigurationHash:        "",
			}

			//TODO: use a real state builder

			mockStateBuilderFactory.EXPECT().NewBuilder(mockBlobstore, mockAgentClient).Return(mockStateBuilder).AnyTimes()
			mockStateBuilder.EXPECT().Build(jobName, jobIndex, gomock.Any(), gomock.Any()).Return(mockState, nil).AnyTimes()
			mockState.EXPECT().ToApplySpec().Return(applySpec).AnyTimes()
		}

		var newDeployCmd = func() Cmd {
			deploymentParser := bideplmanifest.NewParser(fs, logger)
			releaseSetValidator := birelsetmanifest.NewValidator(logger)
			releaseSetParser := birelsetmanifest.NewParser(fs, logger, releaseSetValidator)
			fakeRegistryUUIDGenerator = fakeuuid.NewFakeGenerator()
			fakeRegistryUUIDGenerator.GeneratedUUID = "registry-password"
			installationValidator := biinstallmanifest.NewValidator(logger)
			installationParser := biinstallmanifest.NewParser(fs, fakeRegistryUUIDGenerator, logger, installationValidator)

			deploymentValidator := bideplmanifest.NewValidator(logger)

			instanceFactory := biinstance.NewFactory(mockStateBuilderFactory)
			instanceManagerFactory := biinstance.NewManagerFactory(sshTunnelFactory, instanceFactory, logger)

			pingTimeout := 1 * time.Second
			pingDelay := 100 * time.Millisecond
			deploymentFactory := bidepl.NewFactory(pingTimeout, pingDelay)

			ui := biui.NewWriterUI(stdOut, stdErr, logger)
			doGet := func(deploymentManifestPath string) (DeploymentPreparer, error) {
				// todo: figure this out?
				deploymentStateService = biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath))
				vmRepo = biconfig.NewVMRepo(deploymentStateService)
				diskRepo = biconfig.NewDiskRepo(deploymentStateService, fakeRepoUUIDGenerator)
				stemcellRepo = biconfig.NewStemcellRepo(deploymentStateService, fakeRepoUUIDGenerator)
				deploymentRepo = biconfig.NewDeploymentRepo(deploymentStateService)
				releaseRepo = biconfig.NewReleaseRepo(deploymentStateService, fakeRepoUUIDGenerator)

				legacyDeploymentStateMigrator = biconfig.NewLegacyDeploymentStateMigrator(deploymentStateService, fs, fakeUUIDGenerator, logger)
				deploymentRecord := bidepl.NewRecord(deploymentRepo, releaseRepo, stemcellRepo, fakeSHA1Calculator)
				stemcellManagerFactory = bistemcell.NewManagerFactory(stemcellRepo)
				diskManagerFactory = bidisk.NewManagerFactory(diskRepo, logger)
				diskDeployer = bivm.NewDiskDeployer(diskManagerFactory, diskRepo, logger)
				vmManagerFactory = bivm.NewManagerFactory(vmRepo, stemcellRepo, diskDeployer, fakeAgentIDGenerator, fs, logger)
				deployer := bidepl.NewDeployer(
					vmManagerFactory,
					instanceManagerFactory,
					deploymentFactory,
					logger,
				)
				fakeHTTPClient := fakebihttpclient.NewFakeHTTPClient()
				tarballCache := bitarball.NewCache("fake-base-path", fs, logger)
				tarballProvider := bitarball.NewProvider(tarballCache, fs, fakeHTTPClient, fakeSHA1Calculator, 1, 0, logger)

				cpiInstaller := bicpirel.CpiInstaller{
					ReleaseManager:   releaseManager,
					InstallerFactory: mockInstallerFactory,
					Validator:        bicpirel.NewValidator(),
				}
				releaseFetcher := birel.NewFetcher(tarballProvider, mockReleaseExtractor, releaseManager)
				stemcellFetcher := bistemcell.Fetcher{
					TarballProvider:   tarballProvider,
					StemcellExtractor: fakeStemcellExtractor,
				}

				releaseSetAndInstallationManifestParser := ReleaseSetAndInstallationManifestParser{
					ReleaseSetParser:   releaseSetParser,
					InstallationParser: installationParser,
				}
				deploymentManifestParser := DeploymentManifestParser{
					DeploymentParser:    deploymentParser,
					DeploymentValidator: deploymentValidator,
					ReleaseManager:      releaseManager,
				}

				installationUuidGenerator := fakeuuid.NewFakeGenerator()
				installationUuidGenerator.GeneratedUUID = "fake-installation-id"
				targetProvider := biinstall.NewTargetProvider(
					deploymentStateService,
					installationUuidGenerator,
					filepath.Join("fake-install-dir"),
				)

				tempRootConfigurator := NewTempRootConfigurator(fs)

				return NewDeploymentPreparer(
					ui,
					logger,
					"deployCmd",
					deploymentStateService,
					legacyDeploymentStateMigrator,
					releaseManager,
					deploymentRecord,
					mockCloudFactory,
					stemcellManagerFactory,
					mockAgentClientFactory,
					vmManagerFactory,
					mockBlobstoreFactory,
					deployer,
					deploymentManifestPath,
					cpiInstaller,
					releaseFetcher,
					stemcellFetcher,
					releaseSetAndInstallationManifestParser,
					deploymentManifestParser,
					tempRootConfigurator,
					targetProvider,
				), nil
			}

			return NewDeployCmd(
				ui,
				fs,
				logger,
				doGet,
			)
		}

		var expectDeployFlow = func() {
			agentID := "fake-uuid-0"
			vmCID := "fake-vm-cid-1"
			diskCID := "fake-disk-cid-1"
			diskSize := 1024

			//TODO: use a real StateBuilder and test mockBlobstore.Add & mockAgentClient.CompilePackage

			gomock.InOrder(
				mockCloud.EXPECT().CreateStemcell(stemcellImagePath, stemcellCloudProperties).Return(stemcellCID, nil),
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(vmCID, nil),
				mockCloud.EXPECT().SetVMMetadata(vmCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				mockCloud.EXPECT().CreateDisk(diskSize, diskCloudProperties, vmCID).Return(diskCID, nil),
				mockCloud.EXPECT().AttachDisk(vmCID, diskCID),
				mockAgentClient.EXPECT().MountDisk(diskCID),

				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		var expectDeployWithDiskMigration = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			expectHasVM1 = mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil)

			gomock.InOrder(
				expectHasVM1,

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		var expectDeployWithDiskMigrationMissingVM = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			expectDeleteVM1 = mockCloud.EXPECT().DeleteVM(oldVMCID)

			gomock.InOrder(
				mockCloud.EXPECT().HasVM(oldVMCID).Return(false, nil),

				// delete old vm (without talking to agent) so that the cpi can clean up related resources
				expectDeleteVM1,

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		var expectDeployWithNoDiskToMigrate = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"

			gomock.InOrder(
				mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil),

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attaching a missing disk will fail
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return(
					bicloud.NewCPIError("attach_disk", bicloud.CmdError{
						Type:    bicloud.DiskNotFoundError,
						Message: "fake-disk-not-found-message",
					}),
				),
			)
		}

		var expectDeployWithDiskMigrationFailure = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			gomock.InOrder(
				mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil),

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate (with error)
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk().Return(
					bosherr.Error("fake-migration-error"),
				),
			)
		}

		var expectDeployWithDiskMigrationRepair = func() {
			agentID := "fake-uuid-2"
			oldVMCID := "fake-vm-cid-2"
			newVMCID := "fake-vm-cid-3"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-3"
			newDiskSize := 2048

			gomock.InOrder(
				mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil),

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		var expectRegistryToWork = func() {
			httpClient := bihttp.NewHTTPClient(bihttp.DefaultClient, logger)

			endpoint := "http://registry:registry-password@127.0.0.1:6901/instances/fake-director-id/settings"

			settingsBytes := []byte("fake-registry-contents") //usually json, but not required to be
			response, err := httpClient.Put(endpoint, settingsBytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusCreated))

			response, err = httpClient.Get(endpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
			responseBytes, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(responseBytes).To(Equal([]byte("{\"settings\":\"fake-registry-contents\",\"status\":\"ok\"}")))

			response, err = httpClient.Delete(endpoint)
			Expect(err).ToNot(HaveOccurred())
			Expect(response.StatusCode).To(Equal(http.StatusOK))
		}

		var expectDeployFlowWithRegistry = func() {
			agentID := "fake-uuid-0"
			vmCID := "fake-vm-cid-1"
			diskCID := "fake-disk-cid-1"
			diskSize := 1024

			gomock.InOrder(
				mockCloud.EXPECT().CreateStemcell(stemcellImagePath, stemcellCloudProperties).Do(
					func(_, _ interface{}) { expectRegistryToWork() },
				).Return(stemcellCID, nil),
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Do(
					func(_, _, _, _, _ interface{}) { expectRegistryToWork() },
				).Return(vmCID, nil),
				mockCloud.EXPECT().SetVMMetadata(vmCID, gomock.Any()).Return(nil),

				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				mockCloud.EXPECT().CreateDisk(diskSize, diskCloudProperties, vmCID).Do(
					func(_, _, _ interface{}) { expectRegistryToWork() },
				).Return(diskCID, nil),
				mockCloud.EXPECT().AttachDisk(vmCID, diskCID).Do(
					func(_, _ interface{}) { expectRegistryToWork() },
				),

				mockAgentClient.EXPECT().MountDisk(diskCID),
				mockAgentClient.EXPECT().Stop().Do(
					func() { expectRegistryToWork() },
				),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
			)
		}

		BeforeEach(func() {
			sshConfig = NewSSHConfig(originalHome)
			err := sshConfig.Prepare()
			Expect(err).ToNot(HaveOccurred())

			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()

			logger = boshlog.NewLogger(boshlog.LevelNone)
			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			setupDeploymentStateService := biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath))
			deploymentState, err := setupDeploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			directorID = deploymentState.DirectorID

			fakeAgentIDGenerator = fakeuuid.NewFakeGenerator()

			fakeSHA1Calculator = fakebicrypto.NewFakeSha1Calculator()

			mockInstaller = mock_install.NewMockInstaller(mockCtrl)
			mockInstallerFactory = mock_install.NewMockInstallerFactory(mockCtrl)
			mockCloudFactory = mock_cloud.NewMockFactory(mockCtrl)

			sshTunnelFactory = bisshtunnel.NewFactory(logger)

			fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()

			mockCloud = mock_cloud.NewMockCloud(mockCtrl)

			registryServerManager = biregistry.NewServerManager(logger)

			mockReleaseExtractor = mock_release.NewMockExtractor(mockCtrl)
			releaseManager = birel.NewManager(logger)

			mockStateBuilderFactory = mock_instance_state.NewMockBuilderFactory(mockCtrl)
			mockStateBuilder = mock_instance_state.NewMockBuilder(mockCtrl)
			mockState = mock_instance_state.NewMockState(mockCtrl)

			mockBlobstoreFactory = mock_blobstore.NewMockFactory(mockCtrl)
			mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)
			mockBlobstoreFactory.EXPECT().Create(mbusURL).Return(mockBlobstore, nil).AnyTimes()

			fakeStemcellExtractor = fakebistemcell.NewFakeExtractor()

			stdOut = gbytes.NewBuffer()
			stdErr = gbytes.NewBuffer()
			fakeStage = fakebiui.NewFakeStage()

			mockAgentClientFactory = mock_httpagent.NewMockAgentClientFactory(mockCtrl)
			mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)

			mockAgentClientFactory.EXPECT().NewAgentClient(directorID, mbusURL).Return(mockAgentClient).AnyTimes()

			writeDeploymentManifest()
			writeCPIReleaseTarball()
			writeStemcellReleaseTarball()
		})

		AfterEach(func() {
			err := sshConfig.Restore()
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			allowStemcellToBeExtracted()
			allowCPIToBeInstalled()
			allowApplySpecToBeCreated()
		})

		It("executes the cloud & agent client calls in the expected order", func() {
			expectDeployFlow()

			err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when multiple releases are provided", func() {
			var (
				otherReleaseTarballPath = "/fake-other-release.tgz"
			)

			BeforeEach(func() {
				err := fs.WriteFileString(otherReleaseTarballPath, "fake-other-tgz-content")
				Expect(err).ToNot(HaveOccurred())

				otherRelease := birel.NewRelease(
					"fake-other-release-name",
					"1.2",
					[]bireljob.Job{
						{
							Name: "other",
							Templates: map[string]string{
								"other.erb": "bin/other",
							},
						},
					},
					[]*birelpkg.Package{},
					"fake-other-extracted-dir",
					fs,
				)
				mockReleaseExtractor.EXPECT().Extract(otherReleaseTarballPath).Do(func(_ string) {
					err := fs.MkdirAll("fake-other-extracted-dir", os.ModePerm)
					Expect(err).ToNot(HaveOccurred())
				}).Return(otherRelease, nil).AnyTimes()
			})

			It("extracts all provided releases & finds the cpi release before executing the expected cloud & agent client commands", func() {
				expectDeployFlow()

				err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the deployment state file does not exist", func() {
			BeforeEach(func() {
				err := fs.RemoveAll(deploymentStatePath)
				Expect(err).ToNot(HaveOccurred())

				directorID = "fake-uuid-1"
			})

			It("creates one", func() {
				expectDeployFlow()

				// new directorID will be generated
				mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), mbusURL).Return(mockAgentClient)

				err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists(deploymentStatePath)).To(BeTrue())

				deploymentState, err := deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.DirectorID).To(Equal(directorID))
			})
		})

		Context("when the deployment has been deployed", func() {
			JustBeforeEach(func() {
				expectDeployFlow()

				err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when persistent disk size is increased", func() {
				JustBeforeEach(func() {
					writeDeploymentManifestWithLargerDisk()
				})

				It("migrates the disk content", func() {
					expectDeployWithDiskMigration()

					err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when current VM has been deleted manually (outside of bosh)", func() {
					It("migrates the disk content, but does not shutdown the old VM", func() {
						expectDeployWithDiskMigrationMissingVM()

						err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
						Expect(err).ToNot(HaveOccurred())
					})

					It("ignores DiskNotFound errors", func() {
						expectDeployWithDiskMigrationMissingVM()

						expectDeleteVM1.Return(bicloud.NewCPIError("delete_vm", bicloud.CmdError{
							Type:    bicloud.VMNotFoundError,
							Message: "fake-vm-not-found-message",
						}))

						err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when current disk has been deleted manually (outside of bosh)", func() {
					// because there is no cloud.HasDisk, there is no way to know if the disk does not exist, unless attach/delete fails

					It("returns an error when attach_disk fails with a DiskNotFound error", func() {
						expectDeployWithNoDiskToMigrate()

						err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-disk-not-found-message"))
					})
				})

				Context("after migration has failed", func() {
					JustBeforeEach(func() {
						expectDeployWithDiskMigrationFailure()

						err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migration-error"))

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(HaveLen(2)) // current + unused
					})

					It("deletes unused disks", func() {
						expectDeployWithDiskMigrationRepair()

						mockCloud.EXPECT().DeleteDisk("fake-disk-cid-2")

						err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
						Expect(err).ToNot(HaveOccurred())

						diskRecord, found, err := diskRepo.FindCurrent()
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(BeTrue())
						Expect(diskRecord.CID).To(Equal("fake-disk-cid-3"))

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(Equal([]biconfig.DiskRecord{diskRecord}))
					})
				})
			})

			var expectNoDeployHappened = func() {
				expectDeleteVM := mockCloud.EXPECT().DeleteVM(gomock.Any())
				expectDeleteVM.Times(0)
				expectCreateVM := mockCloud.EXPECT().CreateVM(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
				expectCreateVM.Times(0)

				mockCloud.EXPECT().HasVM(gomock.Any()).Return(true, nil).AnyTimes()
				mockAgentClient.EXPECT().Ping().AnyTimes()
				mockAgentClient.EXPECT().Stop().AnyTimes()
				mockAgentClient.EXPECT().ListDisk().AnyTimes()
			}

			Context("and the same deployment is attempted again", func() {
				It("skips the deploy", func() {
					expectNoDeployHappened()

					err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
					Expect(err).ToNot(HaveOccurred())
					Expect(stdOut).To(gbytes.Say("No deployment, stemcell or release changes. Skipping deploy."))
				})
			})
		})

		Context("when the registry is configured", func() {
			It("makes the registry available for all CPI commands", func() {
				expectDeployFlowWithRegistry()

				err := newDeployCmd().Run(fakeStage, []string{deploymentManifestPath})
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})

type SSHConfig struct {
	Username   string
	PrivateKey string
	HomeDir    string

	rsaKeyDir string
}

func NewSSHConfig(homeDir string) *SSHConfig {
	return &SSHConfig{
		HomeDir: homeDir,
	}
}

func (s *SSHConfig) Prepare() error {
	var err error
	s.Username, err = s.runCommand("whoami")
	if err != nil {
		return err
	}

	s.rsaKeyDir, err = ioutil.TempDir("", "bosh-init-deploy-test")
	if err != nil {
		return err
	}
	s.PrivateKey = filepath.Join(s.rsaKeyDir, "id_rsa_bosh_init")

	script := `
ssh-keygen -f {{ .PrivateKey }} -N ""
mkdir -p {{ .HomeDir }}/.ssh
if [ -f {{ .HomeDir }}/.ssh/authorized_keys ]; then
	cp {{ .HomeDir }}/.ssh/authorized_keys {{ .HomeDir }}/.ssh/authorized_keys_original
fi
cat {{ .PrivateKey }}.pub >> {{ .HomeDir }}/.ssh/authorized_keys
`

	err = s.runScript(script)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSHConfig) Restore() error {
	script := `
if [ -f {{ .HomeDir }}/.ssh/authorized_keys_original ]; then
	mv {{ .HomeDir }}/.ssh/authorized_keys_original {{ .HomeDir }}/.ssh/authorized_keys
else
    rm {{ .HomeDir }}/.ssh/authorized_keys
fi
`
	err := s.runScript(script)
	if err != nil {
		return err
	}

	err = os.RemoveAll(s.rsaKeyDir)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSHConfig) runScript(script string) error {
	buffer := bytes.NewBuffer([]byte{})
	t := template.Must(template.New("manifest").Parse(script))
	t.Execute(buffer, s)
	_, err := s.runCommand("sh", "-c", buffer.String())
	return err
}

func (s *SSHConfig) runCommand(args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	shCmd := exec.Command(args[0], args[1:]...)
	shCmd.Stdout = &stdout
	shCmd.Stderr = &stderr
	if err := shCmd.Run(); err != nil {
		println(stderr.String())
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}
