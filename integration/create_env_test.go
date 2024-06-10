package integration_test

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"
	"time"

	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	mockhttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http/mocks"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/cppforlife/go-patch/patch"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	mockagentclient "github.com/cloudfoundry/bosh-cli/v7/agentclient/mocks"
	mockblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore/mocks"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	mockcloud "github.com/cloudfoundry/bosh-cli/v7/cloud/mocks"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	fakebicrypto "github.com/cloudfoundry/bosh-cli/v7/crypto/fakes"
	bidepl "github.com/cloudfoundry/bosh-cli/v7/deployment"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	mockinstancestate "github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state/mocks"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel"
	bidepltpl "github.com/cloudfoundry/bosh-cli/v7/deployment/template"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	mockinstall "github.com/cloudfoundry/bosh-cli/v7/installation/mocks"
	bitarball "github.com/cloudfoundry/bosh-cli/v7/installation/tarball"
	birel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	birelpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	fakebistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell/stemcellfakes"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("bosh", func() {
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

			releaseManager birel.Manager

			mockInstaller          *mockinstall.MockInstaller
			mockInstallerFactory   *mockinstall.MockInstallerFactory
			mockCloudFactory       *mockcloud.MockFactory
			mockCloud              *mockcloud.MockCloud
			mockAgentClient        *mockagentclient.MockAgentClient
			mockAgentClientFactory *mockhttpagent.MockAgentClientFactory
			releaseReader          *fakerel.FakeReader

			mockStateBuilderFactory *mockinstancestate.MockBuilderFactory
			mockStateBuilder        *mockinstancestate.MockBuilder
			mockState               *mockinstancestate.MockState

			mockBlobstoreFactory *mockblobstore.MockFactory
			mockBlobstore        *mockblobstore.MockBlobstore

			fakeStemcellExtractor         *fakebistemcell.FakeExtractor
			fakeUUIDGenerator             *fakeuuid.FakeGenerator
			fakeRepoUUIDGenerator         *fakeuuid.FakeGenerator
			fakeAgentIDGenerator          *fakeuuid.FakeGenerator
			fakeDigestCalculator          *fakebicrypto.FakeDigestCalculator
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
			deploymentManifestPath = filepath.Join("/", "deployment-dir", "fake-deployment-manifest.yml")
			deploymentStatePath    = filepath.Join("/", "deployment-dir", "fake-deployment-manifest-state.json")

			stemcellCID             = "fake-stemcell-cid"
			stemcellApiVersion      = 2
			cpiApiVersion           = 2
			stemcellCloudProperties = biproperty.Map{}

			vmCloudProperties = biproperty.Map{}
			vmEnv             = biproperty.Map{}

			diskCloudProperties = biproperty.Map{}

			networkInterfaces = map[string]biproperty.Map{
				"network-1": {
					"type":             "dynamic",
					"default":          []bideplmanifest.NetworkDefault{"dns", "gateway"},
					"cloud_properties": biproperty.Map{},
				},
			}

			agentRunningState = biagentclient.AgentState{JobState: "running"}
			mbusURL           = "http://fake-mbus-url"
			caCert            = `-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIQLzf5Fs3v+Dblm+CKQFxiKTANBgkqhkiG9w0BAQsFADAm
MQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkwHhcNMTcwNTE2
MTUzNTI4WhcNMTgwNTE2MTUzNTI4WjAmMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoT
DUNsb3VkIEZvdW5kcnkwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC+
4E0QJMOpQwbHACvrZ4FleP4/DMFvYUBySfKzDOgd99Nm8LdXuJcI1SYHJ3sV+mh0
+cQmRt8U2A/lw7bNU6JdM0fWHa/2nGjSBKWgPzba68NdsmwjqUjLatKpr1yvd384
PJJKC7NrxwvChgB8ui84T4SrXHCioYMDEDIqLGmHJHMKnzQ17nu7ECO4e6QuCfnH
RDs7dTjomTAiFuF4fh4SPgEDMGaCE5HZr4t3gvc9n4UftpcCpi+Jh+neRiWx+v37
ZAYf2kp3wWtYDlgWk06cZzHZZ9uYZFwHDNHdDKHxGGvAh2Rm6rpPF2oA6OEyx6BH
85/STCgSMCnV1Wkd+1yPAgMBAAGjIzAhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMB
Af8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBGvGggx3IM4KCMpVDSv9zFKX4K
IuCRQ6VFab3sgnlelMFaMj3+8baJ/YMko8PP1wVfUviVgKuiZO8tqL00Yo4s1WKp
x3MLIG4eBX9pj0ZVRa3kpcF2Wvg6WhrzUzONf7pfuz/9avl77o4aSt4TwyCvM4Iu
gJ7quVQKcfQcAVwuwWRrZXyhjhHaVKoPP5yRS+ESVTl70J5HBh6B7laooxf1yVAW
8NJK1iQ1Pw2x3ABBo1cSMcTQ3Hk1ZWThJ7oPul2+QyzvOjIjiEPBstyzEPaxPG4I
nH9ttalAwSLBsobVaK8mmiAdtAdx+CmHWrB4UNxCPYasrt5A6a9A9SiQ2dLd
-----END CERTIFICATE-----
`

			expectHasVM1    *gomock.Call
			expectDeleteVM1 *gomock.Call
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
  cert:
    ca: |
      -----BEGIN CERTIFICATE-----
      MIIC+TCCAeGgAwIBAgIQLzf5Fs3v+Dblm+CKQFxiKTANBgkqhkiG9w0BAQsFADAm
      MQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkwHhcNMTcwNTE2
      MTUzNTI4WhcNMTgwNTE2MTUzNTI4WjAmMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoT
      DUNsb3VkIEZvdW5kcnkwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC+
      4E0QJMOpQwbHACvrZ4FleP4/DMFvYUBySfKzDOgd99Nm8LdXuJcI1SYHJ3sV+mh0
      +cQmRt8U2A/lw7bNU6JdM0fWHa/2nGjSBKWgPzba68NdsmwjqUjLatKpr1yvd384
      PJJKC7NrxwvChgB8ui84T4SrXHCioYMDEDIqLGmHJHMKnzQ17nu7ECO4e6QuCfnH
      RDs7dTjomTAiFuF4fh4SPgEDMGaCE5HZr4t3gvc9n4UftpcCpi+Jh+neRiWx+v37
      ZAYf2kp3wWtYDlgWk06cZzHZZ9uYZFwHDNHdDKHxGGvAh2Rm6rpPF2oA6OEyx6BH
      85/STCgSMCnV1Wkd+1yPAgMBAAGjIzAhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMB
      Af8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBGvGggx3IM4KCMpVDSv9zFKX4K
      IuCRQ6VFab3sgnlelMFaMj3+8baJ/YMko8PP1wVfUviVgKuiZO8tqL00Yo4s1WKp
      x3MLIG4eBX9pj0ZVRa3kpcF2Wvg6WhrzUzONf7pfuz/9avl77o4aSt4TwyCvM4Iu
      gJ7quVQKcfQcAVwuwWRrZXyhjhHaVKoPP5yRS+ESVTl70J5HBh6B7laooxf1yVAW
      8NJK1iQ1Pw2x3ABBo1cSMcTQ3Hk1ZWThJ7oPul2+QyzvOjIjiEPBstyzEPaxPG4I
      nH9ttalAwSLBsobVaK8mmiAdtAdx+CmHWrB4UNxCPYasrt5A6a9A9SiQ2dLd
      -----END CERTIFICATE-----
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
				DiskSize: 1024,
			}
			updateManifest(context)

			fakeDigestCalculator.SetCalculateBehavior(map[string]fakebicrypto.CalculateInput{
				deploymentManifestPath: {DigestStr: "fake-deployment-sha1-1"},
			})
		}

		var writeDeploymentManifestWithLargerDisk = func() {
			context := manifestContext{
				DiskSize: 2048,
			}
			updateManifest(context)

			fakeDigestCalculator.SetCalculateBehavior(map[string]fakebicrypto.CalculateInput{
				deploymentManifestPath: {DigestStr: "fake-deployment-sha1-2"},
			})
		}

		var writeCPIReleaseTarball = func() {
			err := fs.WriteFileString(filepath.Join("/", "fake-cpi-release.tgz"), "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowCPIToBeInstalled = func() {
			cpiPackage := birelpkg.NewPackage(NewResource("fake-package-name", "fake-package-fingerprint-cpi", nil), nil)
			job := bireljob.NewJob(NewResource("fake-cpi-release-job-name", "", nil))
			job.Templates = map[string]string{filepath.Join("templates", "cpi.erb"): "bin/cpi"}
			job.PackageNames = []string{"fake-package-name"}
			err := job.AttachPackages([]*birelpkg.Package{cpiPackage})
			Expect(err).ToNot(HaveOccurred())
			cpiRelease := birel.NewRelease(
				"fake-cpi-release-name",
				"1.1",
				"commit",
				false,
				[]*bireljob.Job{job},
				[]*birelpkg.Package{cpiPackage},
				nil,
				nil,
				"fake-cpi-extracted-dir",
				fs,
			)
			releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("/fake-cpi-release.tgz"))
				err := fs.MkdirAll("fake-cpi-extracted-dir", os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
				return cpiRelease, nil
			}

			installationManifest := biinstallmanifest.Manifest{
				Name: "test-deployment",
				Template: biinstallmanifest.ReleaseJobRef{
					Name:    "fake-cpi-release-job-name",
					Release: "fake-cpi-release-name",
				},
				Mbus: mbusURL,
				Cert: biinstallmanifest.Certificate{
					CA: caCert,
				},
				Properties: biproperty.Map{},
			}
			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := biinstall.NewTarget(installationPath)

			installedJob := biinstall.InstalledJob{}
			installedJob.Name = "fake-cpi-release-job-name"
			installedJob.Path = filepath.Join(target.JobsPath(), "fake-cpi-release-job-name")

			installation := biinstall.NewInstallation(target, installedJob, installationManifest)

			mockInstallerFactory.EXPECT().NewInstaller(target).Return(mockInstaller).AnyTimes()

			mockInstaller.EXPECT().Install(installationManifest, gomock.Any()).Do(func(_ interface{}, stage biui.Stage) {
				Expect(fakeStage.SubStages).To(ContainElement(stage))
			}).Return(installation, nil).AnyTimes()
			mockInstaller.EXPECT().Cleanup(installation).AnyTimes()
			mockCloudFactory.EXPECT().NewCloud(installation, directorID, stemcellApiVersion).Return(mockCloud, nil).AnyTimes()
		}

		var writeStemcellReleaseTarball = func() {
			err := fs.WriteFileString(stemcellTarballPath, "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowStemcellToBeExtracted = func() {
			stemcellManifest := bistemcell.Manifest{
				Name:            "fake-stemcell-name",
				Version:         "fake-stemcell-version",
				SHA1:            "fake-stemcell-sha1",
				CloudProperties: biproperty.Map{},
			}

			extractedStemcell := bistemcell.NewExtractedStemcell(
				stemcellManifest,
				"fake-stemcell-extracted-dir",
				fakes.NewFakeCompressor(),
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
				Networks: map[string]interface{}{
					"network-1": map[string]interface{}{
						"cloud_properties": map[string]interface{}{},
						"type":             "dynamic",
						"ip":               "",
					},
				},
				Job: bias.Job{
					Name:      jobName,
					Templates: []bias.Blob{},
				},
				Packages: map[string]bias.Blob{
					"fake-package-name": {
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
			mockStateBuilder.EXPECT().Build(jobName, jobIndex, gomock.Any(), gomock.Any(), gomock.Any()).Return(mockState, nil).AnyTimes()
			mockStateBuilder.EXPECT().BuildInitialState(jobName, jobIndex, gomock.Any()).Return(mockState, nil).AnyTimes()
			mockState.EXPECT().ToApplySpec().Return(applySpec).AnyTimes()
		}

		var newCreateEnvCmd = func() *CreateEnvCmd {
			deploymentParser := bideplmanifest.NewParser(fs, logger)
			releaseSetValidator := birelsetmanifest.NewValidator(logger)
			releaseSetParser := birelsetmanifest.NewParser(fs, logger, releaseSetValidator)
			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			fakeUUIDGenerator.GeneratedUUID = "fake-uuid-1"
			installationValidator := biinstallmanifest.NewValidator(logger)
			installationParser := biinstallmanifest.NewParser(fs, fakeUUIDGenerator, logger, installationValidator)

			deploymentValidator := bideplmanifest.NewValidator(logger)

			instanceFactory := biinstance.NewFactory(mockStateBuilderFactory)
			instanceManagerFactory := biinstance.NewManagerFactory(sshTunnelFactory, instanceFactory, logger)

			pingTimeout := 1 * time.Second
			pingDelay := 100 * time.Millisecond
			deploymentFactory := bidepl.NewFactory(pingTimeout, pingDelay)

			ui := biui.NewWriterUI(stdOut, stdErr, logger)
			doGet := func(deploymentManifestPath string, statePath string, deploymentVars boshtpl.Variables, deploymentOp patch.Op) DeploymentPreparer {
				// todo: figure this out?
				deploymentStateService = biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath, statePath))
				vmRepo = biconfig.NewVMRepo(deploymentStateService)
				diskRepo = biconfig.NewDiskRepo(deploymentStateService, fakeRepoUUIDGenerator)
				stemcellRepo = biconfig.NewStemcellRepo(deploymentStateService, fakeRepoUUIDGenerator)
				deploymentRepo = biconfig.NewDeploymentRepo(deploymentStateService)
				releaseRepo = biconfig.NewReleaseRepo(deploymentStateService, fakeRepoUUIDGenerator)

				legacyDeploymentStateMigrator = biconfig.NewLegacyDeploymentStateMigrator(deploymentStateService, fs, fakeUUIDGenerator, logger)
				deploymentRecord := bidepl.NewRecord(deploymentRepo, releaseRepo, stemcellRepo)
				stemcellManagerFactory = bistemcell.NewManagerFactory(stemcellRepo)
				diskManagerFactory = bidisk.NewManagerFactory(diskRepo, logger)
				diskDeployer = bivm.NewDiskDeployer(diskManagerFactory, diskRepo, logger, false)
				vmManagerFactory = bivm.NewManagerFactory(vmRepo, stemcellRepo, diskDeployer, fakeAgentIDGenerator, fs, logger)
				deployer := bidepl.NewDeployer(
					vmManagerFactory,
					instanceManagerFactory,
					deploymentFactory,
					logger,
				)
				tarballCache := bitarball.NewCache("fake-base-path", fs, logger)
				tarballProvider := bitarball.NewProvider(tarballCache, fs, nil, 1, 0, logger)

				cpiInstaller := bicpirel.CpiInstaller{
					ReleaseManager:   releaseManager,
					InstallerFactory: mockInstallerFactory,
					Validator:        bicpirel.NewValidator(),
				}
				releaseFetcher := biinstall.NewReleaseFetcher(tarballProvider, releaseReader, releaseManager)
				stemcellFetcher := bistemcell.Fetcher{
					TarballProvider:   tarballProvider,
					StemcellExtractor: fakeStemcellExtractor,
				}

				releaseSetAndInstallationManifestParser := ReleaseSetAndInstallationManifestParser{
					ReleaseSetParser:   releaseSetParser,
					InstallationParser: installationParser,
				}
				deploymentManifestParser := NewDeploymentManifestParser(
					deploymentParser,
					deploymentValidator,
					releaseManager,
					bidepltpl.NewDeploymentTemplateFactory(fs),
				)

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
					deploymentVars,
					deploymentOp,
					cpiInstaller,
					releaseFetcher,
					stemcellFetcher,
					releaseSetAndInstallationManifestParser,
					deploymentManifestParser,
					tempRootConfigurator,
					targetProvider,
				)
			}

			return NewCreateEnvCmd(ui, doGet)
		}

		var expectDeployFlow = func() {
			agentID := "fake-uuid-0"
			vmCID := "fake-vm-cid-1"
			diskCID := "fake-disk-cid-1"
			diskSize := 1024

			//TODO: use a real StateBuilder and test mockBlobstore.Add & mockAgentClient.CompilePackage

			gomock.InOrder(
				mockCloud.EXPECT().Info().Return(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil).AnyTimes(),
				mockCloud.EXPECT().CreateStemcell(filepath.Join("fake-stemcell-extracted-dir", "image"), stemcellCloudProperties).Return(stemcellCID, nil),
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(vmCID, nil),
				mockCloud.EXPECT().SetVMMetadata(vmCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				mockCloud.EXPECT().CreateDisk(diskSize, diskCloudProperties, vmCID).Return(diskCID, nil),
				mockCloud.EXPECT().AttachDisk(vmCID, diskCID).Return("/dev/xyz", nil),
				mockCloud.EXPECT().SetDiskMetadata(diskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(diskCID, "/dev/xyz"),
				mockAgentClient.EXPECT().MountDisk(diskCID),

				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().GetState(),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().RunScript("pre-start", map[string]interface{}{}),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
				mockAgentClient.EXPECT().RunScript("post-start", map[string]interface{}{}),
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
				mockCloud.EXPECT().Info().Return(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil),
				expectHasVM1,

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().RunScript("pre-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().Drain("shutdown"),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().RunScript("post-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return("/dev/xyz", nil),
				mockCloud.EXPECT().SetDiskMetadata(oldDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(oldDiskCID, "/dev/xyz"),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID).Return("/dev/abc", nil),
				mockCloud.EXPECT().SetDiskMetadata(newDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(newDiskCID, "/dev/abc"),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockAgentClient.EXPECT().RemovePersistentDisk(oldDiskCID),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().GetState(),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().RunScript("pre-start", map[string]interface{}{}),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
				mockAgentClient.EXPECT().RunScript("post-start", map[string]interface{}{}),
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
				mockCloud.EXPECT().Info().Return(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil),
				mockCloud.EXPECT().HasVM(oldVMCID).Return(false, nil),

				// delete old vm (without talking to agent) so that the cpi can clean up related resources
				expectDeleteVM1,

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return("/dev/xyz", nil),
				mockCloud.EXPECT().SetDiskMetadata(oldDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(oldDiskCID, "/dev/xyz"),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID).Return("/dev/abc", nil),
				mockCloud.EXPECT().SetDiskMetadata(newDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(newDiskCID, "/dev/abc"),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockAgentClient.EXPECT().RemovePersistentDisk(oldDiskCID),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().GetState(),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().RunScript("pre-start", map[string]interface{}{}),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
				mockAgentClient.EXPECT().RunScript("post-start", map[string]interface{}{}),
			)
		}

		var expectDeployWithNoDiskToMigrate = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"

			gomock.InOrder(
				mockCloud.EXPECT().Info().Return(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil),
				mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil),

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().RunScript("pre-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().Drain("shutdown"),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().RunScript("post-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attaching a missing disk will fail
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return(
					"",
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
				mockCloud.EXPECT().Info().Return(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil),
				mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil),

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().RunScript("pre-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().Drain("shutdown"),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().RunScript("post-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				// create new vm
				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate (with error)
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return("/dev/xyz", nil),
				mockCloud.EXPECT().SetDiskMetadata(oldDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(oldDiskCID, "/dev/xyz"),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID).Return("/dev/abc", nil),
				mockCloud.EXPECT().SetDiskMetadata(newDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(newDiskCID, "/dev/abc"),
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
				mockCloud.EXPECT().Info().Return(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil),
				mockCloud.EXPECT().HasVM(oldVMCID).Return(true, nil),

				// shutdown old vm
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().RunScript("pre-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().Drain("shutdown"),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().RunScript("post-stop", map[string]interface{}{}),
				mockAgentClient.EXPECT().ListDisk().Return([]string{oldDiskCID}, nil),
				mockAgentClient.EXPECT().UnmountDisk(oldDiskCID),
				mockCloud.EXPECT().DeleteVM(oldVMCID),

				mockCloud.EXPECT().CreateVM(agentID, stemcellCID, vmCloudProperties, networkInterfaces, vmEnv).Return(newVMCID, nil),
				mockCloud.EXPECT().SetVMMetadata(newVMCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),

				// attach both disks and migrate
				mockCloud.EXPECT().AttachDisk(newVMCID, oldDiskCID).Return("/dev/xyz", nil),
				mockCloud.EXPECT().SetDiskMetadata(oldDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(oldDiskCID, "/dev/xyz"),
				mockAgentClient.EXPECT().MountDisk(oldDiskCID),
				mockCloud.EXPECT().CreateDisk(newDiskSize, diskCloudProperties, newVMCID).Return(newDiskCID, nil),
				mockCloud.EXPECT().AttachDisk(newVMCID, newDiskCID).Return("/dev/abc", nil),
				mockCloud.EXPECT().SetDiskMetadata(newDiskCID, gomock.Any()).Return(nil),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockAgentClient.EXPECT().AddPersistentDisk(newDiskCID, "/dev/abc"),
				mockAgentClient.EXPECT().MountDisk(newDiskCID),
				mockAgentClient.EXPECT().MigrateDisk(),
				mockAgentClient.EXPECT().RemovePersistentDisk(oldDiskCID),
				mockCloud.EXPECT().DetachDisk(newVMCID, oldDiskCID),
				mockAgentClient.EXPECT().Ping().Return("any-state", nil),
				mockCloud.EXPECT().DeleteDisk(oldDiskCID),

				// start jobs & wait for running
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().GetState(),
				mockAgentClient.EXPECT().Stop(),
				mockAgentClient.EXPECT().Apply(applySpec),
				mockAgentClient.EXPECT().RunScript("pre-start", map[string]interface{}{}),
				mockAgentClient.EXPECT().Start(),
				mockAgentClient.EXPECT().GetState().Return(agentRunningState, nil),
				mockAgentClient.EXPECT().RunScript("post-start", map[string]interface{}{}),
			)
		}

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()

			logger = boshlog.NewLogger(boshlog.LevelNone)
			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			setupDeploymentStateService := biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath, ""))
			deploymentState, err := setupDeploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			directorID = deploymentState.DirectorID

			fakeAgentIDGenerator = fakeuuid.NewFakeGenerator()

			fakeDigestCalculator = fakebicrypto.NewFakeDigestCalculator()

			mockInstaller = mockinstall.NewMockInstaller(mockCtrl)
			mockInstallerFactory = mockinstall.NewMockInstallerFactory(mockCtrl)
			mockCloudFactory = mockcloud.NewMockFactory(mockCtrl)

			sshTunnelFactory = bisshtunnel.NewFactory(logger)

			fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()

			mockCloud = mockcloud.NewMockCloud(mockCtrl)

			releaseReader = &fakerel.FakeReader{}
			releaseManager = biinstall.NewReleaseManager(logger)

			mockStateBuilderFactory = mockinstancestate.NewMockBuilderFactory(mockCtrl)
			mockStateBuilder = mockinstancestate.NewMockBuilder(mockCtrl)
			mockState = mockinstancestate.NewMockState(mockCtrl)

			mockBlobstoreFactory = mockblobstore.NewMockFactory(mockCtrl)
			mockBlobstore = mockblobstore.NewMockBlobstore(mockCtrl)
			mockBlobstoreFactory.EXPECT().Create(mbusURL, gomock.Any()).Return(mockBlobstore, nil).AnyTimes()

			fakeStemcellExtractor = fakebistemcell.NewFakeExtractor()

			stdOut = gbytes.NewBuffer()
			stdErr = gbytes.NewBuffer()
			fakeStage = fakebiui.NewFakeStage()

			mockAgentClientFactory = mockhttpagent.NewMockAgentClientFactory(mockCtrl)
			mockAgentClient = mockagentclient.NewMockAgentClient(mockCtrl)

			mockAgentClientFactory.EXPECT().NewAgentClient(directorID, mbusURL, caCert).Return(mockAgentClient, nil).AnyTimes()

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

			err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when multiple releases are provided", func() {
			var (
				otherReleaseTarballPath = filepath.Join("/", "fake-other-release.tgz")
			)

			BeforeEach(func() {
				err := fs.WriteFileString(otherReleaseTarballPath, "fake-other-tgz-content")
				Expect(err).ToNot(HaveOccurred())

				job := bireljob.NewJob(NewResource("other", "", nil))

				otherRelease := birel.NewRelease(
					"fake-other-release-name",
					"1.2",
					"commit",
					false,
					[]*bireljob.Job{job},
					[]*birelpkg.Package{},
					nil,
					nil,
					"fake-other-extracted-dir",
					fs,
				)
				releaseReader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal(otherReleaseTarballPath))
					err := fs.MkdirAll("fake-other-extracted-dir", os.ModePerm)
					Expect(err).ToNot(HaveOccurred())
					return otherRelease, nil
				}
			})

			It("extracts all provided releases & finds the cpi release before executing the expected cloud & agent client commands", func() {
				expectDeployFlow()

				err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the deployment state file does not exist", func() {
			createsStatePath := func(statePath string, createdStatePath string) {
				expectDeployFlow()

				// new directorID will be generated
				mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), mbusURL, caCert).Return(mockAgentClient, nil)

				err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, statePath))
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists(createdStatePath)).To(BeTrue())

				deploymentState, err := deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.DirectorID).To(Equal(directorID))
			}

			Context("and it's NOT specified", func() {
				BeforeEach(func() {
					err := fs.RemoveAll(deploymentStatePath)
					Expect(err).ToNot(HaveOccurred())

					directorID = "fake-uuid-1"
				})

				It("creates one", func() {
					createsStatePath("", deploymentStatePath)
				})
			})

			Context("and it's specified", func() {
				BeforeEach(func() {
					err := fs.RemoveAll(filepath.Join("/", "tmp", "new", "state", "path", "state"))
					Expect(err).ToNot(HaveOccurred())

					directorID = "fake-uuid-1"
				})

				It("creates one", func() {
					createsStatePath(filepath.Join("/", "tmp", "new", "state", "path", "state"), filepath.Join("/", "tmp", "new", "state", "path", "state"))
				})
			})
		})

		Context("when the deployment has been deployed", func() {
			JustBeforeEach(func() {
				expectDeployFlow()

				err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when persistent disk size is increased", func() {
				JustBeforeEach(func() {
					writeDeploymentManifestWithLargerDisk()
				})

				It("migrates the disk content", func() {
					expectDeployWithDiskMigration()

					err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
					Expect(err).ToNot(HaveOccurred())
				})

				Context("when current VM has been deleted manually (outside of bosh)", func() {
					It("migrates the disk content, but does not shutdown the old VM", func() {
						expectDeployWithDiskMigrationMissingVM()

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
						Expect(err).ToNot(HaveOccurred())
					})

					It("ignores DiskNotFound errors", func() {
						expectDeployWithDiskMigrationMissingVM()

						expectDeleteVM1.Return(bicloud.NewCPIError("delete_vm", bicloud.CmdError{
							Type:    bicloud.VMNotFoundError,
							Message: "fake-vm-not-found-message",
						}))

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when current disk has been deleted manually (outside of bosh)", func() {
					// because there is no cloud.HasDisk, there is no way to know if the disk does not exist, unless attach/delete fails

					It("returns an error when attach_disk fails with a DiskNotFound error", func() {
						expectDeployWithNoDiskToMigrate()

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-disk-not-found-message"))
					})
				})

				Context("after migration has failed", func() {
					JustBeforeEach(func() {
						expectDeployWithDiskMigrationFailure()

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migration-error"))

						diskRecords, err := diskRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(diskRecords).To(HaveLen(2)) // current + unused
					})

					It("deletes unused disks", func() {
						expectDeployWithDiskMigrationRepair()

						mockCloud.EXPECT().DeleteDisk("fake-disk-cid-2")

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
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

					err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
					Expect(err).ToNot(HaveOccurred())
					Expect(stdOut).To(gbytes.Say("No deployment, stemcell or release changes. Skipping deploy."))
				})
			})
		})

		Context("when the stemcell supports api_version 2", func() {
			stateFilePath := filepath.Join("/", "tmp", "new", "state", "path", "state")
			stemcellApiVersion = 2
			allowStemcellToBeExtracted = func() {
				stemcellManifest := bistemcell.Manifest{
					Name:            "fake-stemcell-name",
					Version:         "fake-stemcell-version",
					SHA1:            "fake-stemcell-sha1",
					ApiVersion:      stemcellApiVersion,
					CloudProperties: biproperty.Map{},
				}

				extractedStemcell := bistemcell.NewExtractedStemcell(
					stemcellManifest,
					"fake-stemcell-extracted-dir",
					fakes.NewFakeCompressor(),
					fs,
				)
				fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, extractedStemcell, nil)
			}

			BeforeEach(func() {
				err := fs.RemoveAll(stateFilePath)
				Expect(err).ToNot(HaveOccurred())

				directorID = "fake-uuid-1"
			})

			It("uses the version with the cpi api calls", func() {
				expectDeployFlow()

				// new directorID will be generated
				mockCloudFactory.EXPECT().NewCloud(gomock.Any(), directorID, stemcellApiVersion).Return(mockCloud, nil).AnyTimes()
				mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), mbusURL, caCert).Return(mockAgentClient, nil)

				err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, stateFilePath))
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists(stateFilePath)).To(BeTrue())

				deploymentState, err := deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.Stemcells[0].ApiVersion).To(Equal(2))
			})
		})
	})
})

func newDeployOpts(manifestPath string, statePath string) CreateEnvOpts {
	return CreateEnvOpts{StatePath: statePath, Args: CreateEnvArgs{Manifest: FileBytesWithPathArg{Path: manifestPath}}}
}
