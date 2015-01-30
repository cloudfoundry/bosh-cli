package integration_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_cloud "github.com/cloudfoundry/bosh-micro-cli/cloud/mocks"
	mock_httpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"
	mock_instance_state "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state/mocks"
	mock_install "github.com/cloudfoundry/bosh-micro-cli/installation/mocks"
	mock_release "github.com/cloudfoundry/bosh-micro-cli/release/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmhttp "github.com/cloudfoundry/bosh-micro-cli/deployment/httpclient"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"

	fakebmcrypto "github.com/cloudfoundry/bosh-micro-cli/crypto/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("bosh-micro", func() {
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

			registryServerManager bmregistry.ServerManager
			releaseManager        bmrel.Manager
			releaseResolver       bmrelset.Resolver

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

			fakeStemcellExtractor   *fakebmstemcell.FakeExtractor
			fakeUUIDGenerator       *fakeuuid.FakeGenerator
			fakeRepoUUIDGenerator   *fakeuuid.FakeGenerator
			fakeAgentIDGenerator    *fakeuuid.FakeGenerator
			fakeSHA1Calculator      *fakebmcrypto.FakeSha1Calculator
			deploymentConfigService bmconfig.DeploymentConfigService
			vmRepo                  bmconfig.VMRepo
			diskRepo                bmconfig.DiskRepo
			stemcellRepo            bmconfig.StemcellRepo
			deploymentRepo          bmconfig.DeploymentRepo
			releaseRepo             bmconfig.ReleaseRepo
			userConfig              bmconfig.UserConfig

			sshTunnelFactory bmsshtunnel.Factory

			diskManagerFactory bmdisk.ManagerFactory
			diskDeployer       bmvm.DiskDeployer

			ui          *fakeui.FakeUI
			eventLogger bmeventlog.EventLogger

			stemcellManagerFactory bmstemcell.ManagerFactory
			vmManagerFactory       bmvm.ManagerFactory

			applySpec bmas.ApplySpec

			directorID string

			stemcellTarballPath    = "/fake-stemcell-release.tgz"
			cpiReleaseTarballPath  = "/fake-cpi-release.tgz"
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			deploymentConfigPath   = "/fake-bosh-deployments.json"

			cloudProperties   = map[string]interface{}{}
			stemcellImagePath = "fake-stemcell-image-path"
			stemcellCID       = "fake-stemcell-cid"
			env               = map[string]interface{}{}
			networkInterfaces = map[string]map[string]interface{}{
				"network-1": map[string]interface{}{
					"type":             "dynamic",
					"ip":               "",
					"cloud_properties": cloudProperties,
				},
			}
			agentRunningState = bmagentclient.AgentState{JobState: "running"}
			mbusURL           = "http://fake-mbus-url"

			expectHasVM1    *gomock.Call
			expectDeleteVM1 *gomock.Call
		)

		var writeDeploymentManifest = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---
name: test-release

releases:
- name: fake-cpi-release-name
  version: 1.1

networks:
- name: network-1
  type: dynamic

resource_pools:
- name: resource-pool-1
  network: network-1

jobs:
- name: cpi
  instances: 1
  persistent_disk: 1024
  networks:
  - name: network-1
  templates:
  - {name: cpi, release: fake-cpi-release-name}

cloud_provider:
  release: fake-cpi-release-name
  mbus: http://fake-mbus-url
  registry:
    host: 127.0.0.1
    port: 6301
    username: fake-registry-user
    password: fake-registry-password
`)
			Expect(err).ToNot(HaveOccurred())

			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
				deploymentManifestPath: {Sha1: "fake-deployment-sha1-1"},
			})
		}

		var writeDeploymentManifestWithLargerDisk = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---
name: test-release

releases:
- name: fake-cpi-release-name
  version: 1.1

networks:
- name: network-1
  type: dynamic

resource_pools:
- name: resource-pool-1
  network: network-1

jobs:
- name: cpi
  instances: 1
  persistent_disk: 2048
  networks:
  - name: network-1
  templates:
  - {name: cpi, release: fake-cpi-release-name}

cloud_provider:
  release: fake-cpi-release-name
  mbus: http://fake-mbus-url
  registry:
    host: 127.0.0.1
    port: 6301
    username: fake-registry-user
    password: fake-registry-password
`)
			Expect(err).ToNot(HaveOccurred())

			fakeSHA1Calculator.SetCalculateBehavior(map[string]fakebmcrypto.CalculateInput{
				deploymentManifestPath: {Sha1: "fake-deployment-sha1-2"},
			})
		}

		var writeCPIReleaseTarball = func() {
			err := fs.WriteFileString(cpiReleaseTarballPath, "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowCPIToBeInstalled = func() {
			cpiPackage := bmrel.Package{
				Name:          "cpi",
				Fingerprint:   "fake-package-fingerprint-cpi",
				SHA1:          "fake-package-sha1-cpi",
				Dependencies:  []*bmrel.Package{},
				ExtractedPath: "fake-package-extracted-path-cpi",
				ArchivePath:   "fake-package-archive-path-cpi",
			}
			cpiRelease := bmrel.NewRelease(
				"fake-cpi-release-name",
				"1.1",
				[]bmrel.Job{
					{
						Name: "cpi",
						Templates: map[string]string{
							"cpi.erb": "bin/cpi",
						},
						Packages: []*bmrel.Package{&cpiPackage},
					},
				},
				[]*bmrel.Package{&cpiPackage},
				"fake-cpi-extracted-dir",
				fs,
			)
			mockReleaseExtractor.EXPECT().Extract(cpiReleaseTarballPath).Do(func(_ string) {
				err := fs.MkdirAll("fake-cpi-extracted-dir", os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
			}).Return(cpiRelease, nil).AnyTimes()

			installationManifest := bminstallmanifest.Manifest{
				Name:    "test-release",
				Release: "fake-cpi-release-name",
				Mbus:    mbusURL,
				Registry: bminstallmanifest.Registry{
					Username: "fake-registry-user",
					Password: "fake-registry-password",
					Host:     "127.0.0.1",
					Port:     6301,
				},
			}

			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := bminstall.NewTarget(installationPath)

			installedJob := bminstalljob.InstalledJob{
				Name: "cpi",
				Path: filepath.Join(target.JobsPath(), "cpi"),
			}

			installation := bminstall.NewInstallation(target, installedJob, installationManifest, registryServerManager)

			mockInstallerFactory.EXPECT().NewInstaller().Return(mockInstaller, nil).AnyTimes()
			mockInstaller.EXPECT().Install(installationManifest).Return(installation, nil).AnyTimes()

			mockCloudFactory.EXPECT().NewCloud(installation, directorID).Return(mockCloud, nil).AnyTimes()
		}

		var writeStemcellReleaseTarball = func() {
			err := fs.WriteFileString(stemcellTarballPath, "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowStemcellToBeExtracted = func() {
			stemcellManifest := bmstemcell.Manifest{
				ImagePath: "fake-stemcell-image-path",
				Name:      "fake-stemcell-name",
				Version:   "fake-stemcell-version",
				SHA1:      "fake-stemcell-sha1",
			}
			stemcellApplySpec := bmstemcell.ApplySpec{
				Job: bmstemcell.Job{
					Name:      "cpi",
					Templates: []bmstemcell.Blob{},
				},
				Packages: map[string]bmstemcell.Blob{},
				Networks: map[string]interface{}{},
			}
			extractedStemcell := bmstemcell.NewExtractedStemcell(
				stemcellManifest,
				stemcellApplySpec,
				"fake-stemcell-extracted-dir",
				fs,
			)
			fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, extractedStemcell, nil)
		}

		var allowApplySpecToBeCreated = func() {
			jobName := "cpi"
			jobIndex := 0

			applySpec = bmas.ApplySpec{
				Deployment: "test-release",
				Index:      jobIndex,
				Networks: map[string]interface{}{
					"network-1": map[string]interface{}{
						"cloud_properties": map[string]interface{}{},
						"type":             "dynamic",
						"ip":               "",
					},
				},
				Job: bmas.Job{
					Name:      jobName,
					Templates: []bmas.Blob{},
				},
				Packages: map[string]bmas.Blob{
					"cpi": bmas.Blob{
						Name:        "cpi",
						Version:     "fake-package-fingerprint-cpi",
						SHA1:        "fake-compiled-package-sha1-cpi",
						BlobstoreID: "fake-compiled-package-blob-id-cpi",
					},
				},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{},
				ConfigurationHash:        "",
			}

			//TODO: use a real state builder

			mockStateBuilderFactory.EXPECT().NewBuilder(mockBlobstore, mockAgentClient).Return(mockStateBuilder).AnyTimes()
			mockStateBuilder.EXPECT().Build(jobName, jobIndex, gomock.Any(), gomock.Any()).Return(mockState, nil).AnyTimes()
			mockState.EXPECT().ToApplySpec().Return(applySpec).AnyTimes()
		}

		var newDeployCmd = func() Cmd {
			deploymentParser := bmdeplmanifest.NewParser(fs, logger)
			releaseSetParser := bmrelsetmanifest.NewParser(fs, logger)
			installationParser := bminstallmanifest.NewParser(fs, logger)

			releaseSetValidator := bmrelsetmanifest.NewValidator(logger, releaseResolver)
			installationValidator := bminstallmanifest.NewValidator(logger, releaseResolver)
			deploymentValidator := bmdeplmanifest.NewValidator(logger, releaseResolver)

			deploymentRecord := bmdepl.NewRecord(deploymentRepo, releaseRepo, stemcellRepo, fakeSHA1Calculator)

			instanceFactory := bminstance.NewFactory(mockStateBuilderFactory)
			instanceManagerFactory := bminstance.NewManagerFactory(sshTunnelFactory, instanceFactory, logger)

			pingTimeout := 1 * time.Second
			pingDelay := 100 * time.Millisecond
			deploymentFactory := bmdepl.NewFactory(pingTimeout, pingDelay)

			deployer := bmdepl.NewDeployer(
				vmManagerFactory,
				instanceManagerFactory,
				deploymentFactory,
				eventLogger,
				logger,
			)

			return NewDeployCmd(
				ui,
				userConfig,
				fs,
				releaseSetParser,
				installationParser,
				deploymentParser,
				deploymentConfigService,
				releaseSetValidator,
				installationValidator,
				deploymentValidator,
				mockInstallerFactory,
				mockReleaseExtractor,
				releaseManager,
				releaseResolver,
				mockCloudFactory,
				mockAgentClientFactory,
				vmManagerFactory,
				fakeStemcellExtractor,
				stemcellManagerFactory,
				deploymentRecord,
				mockBlobstoreFactory,
				deployer,
				eventLogger,
				logger,
			)
		}

		var expectDeployFlow = func() {
			agentID := "fake-uuid-0"
			vmCID := "fake-vm-cid-1"
			diskCID := "fake-disk-cid-1"
			diskSize := 1024

			//TODO: use a real StateBuilder and test mockBlobstore.Add & mockAgentClient.CompilePackage

			gomock.InOrder(
				mockCloud.EXPECT().CreateStemcell(stemcellImagePath, cloudProperties).Return(stemcellCID, nil),
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Return(vmCID, nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				mockCloud.EXPECT().CreateDisk(diskSize, cloudProperties, vmCID).Return(diskCID, nil),
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
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Return(newVMCID, nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
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
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Return(newVMCID, nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
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
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Return(newVMCID, nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attaching a missing disk will fail
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return(bmcloud.NewCPIError("attach_disk", bmcloud.CmdError{
					Type:    bmcloud.DiskNotFoundError,
					Message: "fake-disk-not-found-message",
				})),
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
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Return(newVMCID, nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate (with error)
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk().Return(errors.New("fake-migration-error")),
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
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Return(newVMCID, nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, cloudProperties, newVMCID).Return(newDiskCID, nil),
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
			httpClient := bmhttp.NewHTTPClient(logger)

			endpoint := "http://fake-registry-user:fake-registry-password@127.0.0.1:6301/instances/fake-director-id/settings"

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
				mockCloud.EXPECT().CreateStemcell(stemcellImagePath, cloudProperties).Do(
					func(_, _ interface{}) { expectRegistryToWork() },
				).Return(stemcellCID, nil),
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env).Do(
					func(_, _, _, _, _ interface{}) { expectRegistryToWork() },
				).Return(vmCID, nil),

				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				mockCloud.EXPECT().CreateDisk(diskSize, cloudProperties, vmCID).Do(
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
			fs = fakesys.NewFakeFileSystem()
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, fs, fakeUUIDGenerator, logger)
			fakeAgentIDGenerator = fakeuuid.NewFakeGenerator()

			fakeSHA1Calculator = fakebmcrypto.NewFakeSha1Calculator()

			mockInstaller = mock_install.NewMockInstaller(mockCtrl)
			mockInstallerFactory = mock_install.NewMockInstallerFactory(mockCtrl)
			mockCloudFactory = mock_cloud.NewMockFactory(mockCtrl)

			sshTunnelFactory = bmsshtunnel.NewFactory(logger)

			config, err := deploymentConfigService.Load()
			Expect(err).ToNot(HaveOccurred())
			directorID = config.DirectorID

			fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()
			vmRepo = bmconfig.NewVMRepo(deploymentConfigService)
			diskRepo = bmconfig.NewDiskRepo(deploymentConfigService, fakeRepoUUIDGenerator)
			stemcellRepo = bmconfig.NewStemcellRepo(deploymentConfigService, fakeRepoUUIDGenerator)
			deploymentRepo = bmconfig.NewDeploymentRepo(deploymentConfigService)
			releaseRepo = bmconfig.NewReleaseRepo(deploymentConfigService, fakeRepoUUIDGenerator)

			diskManagerFactory = bmdisk.NewManagerFactory(diskRepo, logger)
			diskDeployer = bmvm.NewDiskDeployer(diskManagerFactory, diskRepo, logger)

			mockCloud = mock_cloud.NewMockCloud(mockCtrl)

			registryServerManager = bmregistry.NewServerManager(logger)

			mockReleaseExtractor = mock_release.NewMockExtractor(mockCtrl)
			releaseManager = bmrel.NewManager(logger)
			releaseResolver = bmrelset.NewResolver(releaseManager, logger)

			mockStateBuilderFactory = mock_instance_state.NewMockBuilderFactory(mockCtrl)
			mockStateBuilder = mock_instance_state.NewMockBuilder(mockCtrl)
			mockState = mock_instance_state.NewMockState(mockCtrl)

			mockBlobstoreFactory = mock_blobstore.NewMockFactory(mockCtrl)
			mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)
			mockBlobstoreFactory.EXPECT().Create(mbusURL).Return(mockBlobstore, nil).AnyTimes()

			fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()

			ui = &fakeui.FakeUI{}
			eventLogger = bmeventlog.NewEventLogger(ui)

			mockAgentClientFactory = mock_httpagent.NewMockAgentClientFactory(mockCtrl)
			mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)

			stemcellManagerFactory = bmstemcell.NewManagerFactory(stemcellRepo)

			vmManagerFactory = bmvm.NewManagerFactory(
				vmRepo,
				stemcellRepo,
				diskDeployer,
				fakeAgentIDGenerator,
				fs,
				logger,
			)

			userConfig = bmconfig.UserConfig{DeploymentManifestPath: deploymentManifestPath}

			mockAgentClientFactory.EXPECT().NewAgentClient(directorID, mbusURL).Return(mockAgentClient).AnyTimes()

			writeDeploymentManifest()
			writeCPIReleaseTarball()
			writeStemcellReleaseTarball()
		})

		JustBeforeEach(func() {
			allowStemcellToBeExtracted()
			allowCPIToBeInstalled()
			allowApplySpecToBeCreated()
		})

		It("executes the cloud & agent client calls in the expected order", func() {
			expectDeployFlow()

			err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when multiple releases are provided", func() {
			var (
				otherReleaseTarballPath = "/fake-other-release.tgz"
			)

			BeforeEach(func() {
				err := fs.WriteFileString(otherReleaseTarballPath, "fake-other-tgz-content")
				Expect(err).ToNot(HaveOccurred())

				otherRelease := bmrel.NewRelease(
					"fake-other-release-name",
					"1.2",
					[]bmrel.Job{
						{
							Name: "other",
							Templates: map[string]string{
								"other.erb": "bin/other",
							},
						},
					},
					[]*bmrel.Package{},
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

				err := newDeployCmd().Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the deployment has not been set", func() {
			BeforeEach(func() {
				userConfig.DeploymentManifestPath = ""
			})

			It("returns an error", func() {
				err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Deployment manifest not set"))
			})
		})

		Context("when the deployment config file does not exist", func() {
			BeforeEach(func() {
				err := fs.RemoveAll(deploymentConfigPath)
				Expect(err).ToNot(HaveOccurred())

				directorID = "fake-uuid-1"
			})

			It("creates one", func() {
				expectDeployFlow()

				// new directorID will be generated
				mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), mbusURL).Return(mockAgentClient)

				err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists(deploymentConfigPath)).To(BeTrue())

				deploymentConfig, err := deploymentConfigService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentConfig.DirectorID).To(Equal(directorID))
			})
		})

		Context("when the deployment has been deployed", func() {
			JustBeforeEach(func() {
				expectDeployFlow()

				err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).ToNot(HaveOccurred())

				// reset output buffer
				ui.Said = []string{}
			})

			Context("when persistent disk size is increased", func() {
				JustBeforeEach(func() {
					writeDeploymentManifestWithLargerDisk()
				})

				It("migrates the disk content", func() {
					expectDeployWithDiskMigration()

					err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when current VM has been deleted manually (outside of bosh)", func() {
					It("migrates the disk content, but does not shutdown the old VM", func() {
						expectDeployWithDiskMigrationMissingVM()

						err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
						Expect(err).ToNot(HaveOccurred())
					})

					It("ignores DiskNotFound errors", func() {
						expectDeployWithDiskMigrationMissingVM()

						expectDeleteVM1.Return(bmcloud.NewCPIError("delete_vm", bmcloud.CmdError{
							Type:    bmcloud.VMNotFoundError,
							Message: "fake-vm-not-found-message",
						}))

						err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when current disk has been deleted manually (outside of bosh)", func() {
					// because there is no cloud.HasDisk, there is no way to know if the disk does not exist, unless attach/delete fails

					It("returns an error when attach_disk fails with a DiskNotFound error", func() {
						expectDeployWithNoDiskToMigrate()

						err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-disk-not-found-message"))
					})
				})

				Context("after migration has failed", func() {
					JustBeforeEach(func() {
						expectDeployWithDiskMigrationFailure()

						err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migration-error"))

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(HaveLen(2)) // current + unused

						// reset output buffer
						ui.Said = []string{}
					})

					It("deletes unused disks", func() {
						expectDeployWithDiskMigrationRepair()

						mockCloud.EXPECT().DeleteDisk("fake-disk-cid-2")

						err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
						Expect(err).ToNot(HaveOccurred())

						diskRecord, found, err := diskRepo.FindCurrent()
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(BeTrue())
						Expect(diskRecord.CID).To(Equal("fake-disk-cid-3"))

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(Equal([]bmconfig.DiskRecord{diskRecord}))
					})
				})
			})
		})

		Context("when the registry is configured", func() {
			It("makes the registry available for all CPI commands", func() {
				expectDeployFlowWithRegistry()

				err := newDeployCmd().Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
