package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	biagentclient "github.com/cloudfoundry/bosh-agent/v2/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/v2/agentclient/applyspec"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry/bosh-cli/v7/agentclient/agentclientfakes"
	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	"github.com/cloudfoundry/bosh-cli/v7/crypto/cryptofakes"
	bidepl "github.com/cloudfoundry/bosh-cli/v7/deployment"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/v7/deployment/instance"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/instance/state/statefakes"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel"
	bidepltpl "github.com/cloudfoundry/bosh-cli/v7/deployment/template"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	"github.com/cloudfoundry/bosh-cli/v7/installation/installationfakes"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	bitarball "github.com/cloudfoundry/bosh-cli/v7/installation/tarball"
	birel "github.com/cloudfoundry/bosh-cli/v7/release"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	birelpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	fakebistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell/stemcellfakes"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("bosh", func() {
	Describe("Deploy", func() {
		var (
			fs     *fakesys.FakeFileSystem
			logger boshlog.Logger

			releaseManager birel.Manager

			mockInstaller          *installationfakes.FakeInstaller
			mockInstallerFactory   *installationfakes.FakeInstallerFactory
			mockCloudFactory       *cloudfakes.FakeFactory
			mockCloud              *cloudfakes.FakeCloud
			mockAgentClient        *agentclientfakes.FakeAgentClient
			mockAgentClientFactory *fakecmd.FakeAgentClientFactory
			releaseReader          *fakerel.FakeReader

			mockStateBuilderFactory *statefakes.FakeBuilderFactory
			mockStateBuilder        *statefakes.FakeBuilder
			mockState               *statefakes.FakeState

			mockBlobstoreFactory *blobstorefakes.FakeFactory
			mockBlobstore        *blobstorefakes.FakeBlobstore

			// callOrder records, in invocation order, calls to mockCloud/mockAgentClient
			// methods made during a deploy run -- the counterfeiter equivalent of gomock.InOrder().
			callOrder []string

			fakeStemcellExtractor         *fakebistemcell.FakeExtractor
			fakeUUIDGenerator             *fakeuuid.FakeGenerator
			fakeRepoUUIDGenerator         *fakeuuid.FakeGenerator
			fakeAgentIDGenerator          *fakeuuid.FakeGenerator
			fakeDigestCalculator          *cryptofakes.FakeDigestCalculator
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
			fakeStage *testui.Stage

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

			deleteOldVM1Err error
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
			DiskSize int
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

			fakeDigestCalculator.CalculateStub = func(path string) (string, error) {
				switch path {
				case deploymentManifestPath:
					return "fake-deployment-sha1-1", nil
				default:
					return "", fmt.Errorf("unexpected input '%s'", path)
				}
			}
		}

		var writeDeploymentManifestWithLargerDisk = func() {
			context := manifestContext{
				DiskSize: 2048,
			}
			updateManifest(context)

			fakeDigestCalculator.CalculateStub = func(path string) (string, error) {
				switch path {
				case deploymentManifestPath:
					return "fake-deployment-sha1-2", nil
				default:
					return "", fmt.Errorf("unexpected input '%s'", path)
				}
			}
		}

		var writeCPIReleaseTarball = func() {
			err := fs.WriteFileString(filepath.Join("/", "fake-cpi-release.tgz"), "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowCPIToBeInstalled = func() {
			cpiPackage := birelpkg.NewPackage(resource.NewResource("fake-package-name", "fake-package-fingerprint-cpi", nil), nil)
			job := bireljob.NewJob(resource.NewResource("fake-cpi-release-job-name", "", nil))
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
				false,
				"fake-cpi-extracted-dir",
				fs,
			)
			releaseReader.ReadStub = func(path string) (birel.Release, error) {
				Expect(path).To(Equal("/fake-cpi-release.tgz"))
				err := fs.MkdirAll("fake-cpi-extracted-dir", os.ModePerm)
				Expect(err).ToNot(HaveOccurred())
				return cpiRelease, nil
			}

			installationManifest := biinstallmanifest.Manifest{
				Name: "test-deployment",
				Templates: []biinstallmanifest.ReleaseJobRef{
					{Name: "fake-cpi-release-job-name", Release: "fake-cpi-release-name"},
				},
				Mbus: mbusURL,
				Cert: biinstallmanifest.Certificate{
					CA: caCert,
				},
				Properties: biproperty.Map{},
			}
			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := biinstall.NewTarget(installationPath, "")

			installedJob := biinstall.InstalledJob{}
			installedJob.Name = "fake-cpi-release-job-name"
			installedJob.Path = filepath.Join(target.JobsPath(), "fake-cpi-release-job-name")

			installation := biinstall.NewInstallation(target, []biinstall.InstalledJob{installedJob},
				installationManifest)

			mockInstallerFactory.NewInstallerReturns(mockInstaller)

			mockInstaller.InstallStub = func(_ biinstallmanifest.Manifest, stage biui.Stage) (biinstall.Installation, error) {
				Expect(fakeStage.SubStages).To(ContainElement(stage))
				return installation, nil
			}
			mockInstaller.CleanupReturns(nil)
			mockCloudFactory.NewCloudReturns(mockCloud, nil)
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

			// TODO: use a real state builder

			mockStateBuilderFactory.NewBuilderReturns(mockStateBuilder)
			mockStateBuilder.BuildReturns(mockState, nil)
			mockStateBuilder.BuildInitialStateReturns(mockState, nil)
			mockState.ToApplySpecReturns(applySpec)
		}

		var newCreateEnvCmd = func() *cmd.CreateEnvCmd {
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
			doGet := func(deploymentManifestPath string, statePath string, deploymentVars boshtpl.Variables, deploymentOp patch.Op) cmd.DeploymentPreparer {
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
				}
				releaseFetcher := biinstall.NewReleaseFetcher(tarballProvider, releaseReader, releaseManager)
				stemcellFetcher := bistemcell.Fetcher{
					TarballProvider:   tarballProvider,
					StemcellExtractor: fakeStemcellExtractor,
				}

				releaseSetAndInstallationManifestParser := cmd.ReleaseSetAndInstallationManifestParser{
					ReleaseSetParser:   releaseSetParser,
					InstallationParser: installationParser,
				}
				deploymentManifestParser := cmd.NewDeploymentManifestParser(
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
					"",
				)

				tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

				return cmd.NewDeploymentPreparer(
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

			return cmd.NewCreateEnvCmd(ui, doGet)
		}

		var wireCommonStubs = func() {
			mockCloud.InfoReturns(bicloud.CpiInfo{ApiVersion: cpiApiVersion}, nil)

			mockCloud.SetVMMetadataStub = func(string, bicloud.VMMetadata) error {
				callOrder = append(callOrder, "SetVMMetadata")
				return nil
			}
			mockCloud.SetDiskMetadataStub = func(string, bicloud.DiskMetadata) error {
				callOrder = append(callOrder, "SetDiskMetadata")
				return nil
			}
			mockCloud.DeleteDiskStub = func(string) error {
				callOrder = append(callOrder, "DeleteDisk")
				return nil
			}
			mockCloud.DetachDiskStub = func(string, string) error {
				callOrder = append(callOrder, "DetachDisk")
				return nil
			}

			mockAgentClient.PingStub = func() (string, error) {
				callOrder = append(callOrder, "Ping")
				return "any-state", nil
			}
			mockAgentClient.RunScriptStub = func(name string, _ map[string]interface{}) error {
				callOrder = append(callOrder, "RunScript:"+name)
				return nil
			}
			mockAgentClient.DrainStub = func(string) (int64, error) {
				callOrder = append(callOrder, "Drain")
				return 0, nil
			}
			mockAgentClient.StopStub = func() error {
				callOrder = append(callOrder, "Stop")
				return nil
			}
			mockAgentClient.StartStub = func() error {
				callOrder = append(callOrder, "Start")
				return nil
			}
			mockAgentClient.ApplyStub = func(bias.ApplySpec) error {
				callOrder = append(callOrder, "Apply")
				return nil
			}
			mockAgentClient.UnmountDiskStub = func(string) error {
				callOrder = append(callOrder, "UnmountDisk")
				return nil
			}
			mockAgentClient.AddPersistentDiskStub = func(string, interface{}) error {
				callOrder = append(callOrder, "AddPersistentDisk")
				return nil
			}
			mockAgentClient.MountDiskStub = func(string) error {
				callOrder = append(callOrder, "MountDisk")
				return nil
			}
			mockAgentClient.RemovePersistentDiskStub = func(string) error {
				callOrder = append(callOrder, "RemovePersistentDisk")
				return nil
			}
			mockAgentClient.MigrateDiskStub = func() error {
				callOrder = append(callOrder, "MigrateDisk")
				return nil
			}

			getStateCallCount := 0
			mockAgentClient.GetStateStub = func() (biagentclient.AgentState, error) {
				callOrder = append(callOrder, "GetState")
				getStateCallCount++
				if getStateCallCount >= 2 {
					return agentRunningState, nil
				}
				return biagentclient.AgentState{}, nil
			}
		}

		var expectDeployFlow = func() {
			agentID := "fake-uuid-0"
			vmCID := "fake-vm-cid-1"
			diskCID := "fake-disk-cid-1"
			diskSize := 1024

			// TODO: use a real StateBuilder and test mockBlobstore.Add & mockAgentClient.CompilePackage

			callOrder = nil
			wireCommonStubs()

			mockCloud.CreateStemcellStub = func(imagePath string, cloudProperties biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateStemcell")
				Expect(imagePath).To(Equal(filepath.Join("fake-stemcell-extracted-dir", "image")))
				Expect(cloudProperties).To(Equal(stemcellCloudProperties))
				return stemcellCID, nil
			}
			mockCloud.CreateVMStub = func(gotAgentID, gotStemcellCID string, cloudProperties biproperty.Map, diskCIDs []string, networks map[string]biproperty.Map, env biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateVM")
				Expect(gotAgentID).To(Equal(agentID))
				Expect(gotStemcellCID).To(Equal(stemcellCID))
				Expect(cloudProperties).To(Equal(vmCloudProperties))
				Expect(networks).To(Equal(networkInterfaces))
				Expect(env).To(Equal(vmEnv))
				return vmCID, nil
			}
			mockCloud.CreateDiskStub = func(size int, cloudProperties biproperty.Map, gotVMCID string) (string, error) {
				callOrder = append(callOrder, "CreateDisk")
				Expect(size).To(Equal(diskSize))
				Expect(cloudProperties).To(Equal(diskCloudProperties))
				Expect(gotVMCID).To(Equal(vmCID))
				return diskCID, nil
			}
			mockCloud.AttachDiskStub = func(gotVMCID, gotDiskCID string) (interface{}, error) {
				callOrder = append(callOrder, "AttachDisk")
				Expect(gotVMCID).To(Equal(vmCID))
				Expect(gotDiskCID).To(Equal(diskCID))
				return "/dev/xyz", nil
			}
		}

		var expectDeployWithDiskMigration = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			callOrder = nil
			wireCommonStubs()

			mockCloud.HasVMStub = func(vmCID string) (bool, error) {
				callOrder = append(callOrder, "HasVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return true, nil
			}
			mockCloud.DeleteVMStub = func(vmCID string) error {
				callOrder = append(callOrder, "DeleteVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return nil
			}
			mockAgentClient.ListDiskStub = func() ([]string, error) {
				callOrder = append(callOrder, "ListDisk")
				return []string{oldDiskCID}, nil
			}
			mockCloud.CreateVMStub = func(gotAgentID, gotStemcellCID string, cloudProperties biproperty.Map, diskCIDs []string, networks map[string]biproperty.Map, env biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateVM")
				Expect(gotAgentID).To(Equal(agentID))
				Expect(gotStemcellCID).To(Equal(stemcellCID))
				Expect(diskCIDs).To(Equal([]string{oldDiskCID}))
				return newVMCID, nil
			}
			mockCloud.CreateDiskStub = func(size int, cloudProperties biproperty.Map, vmCID string) (string, error) {
				callOrder = append(callOrder, "CreateDisk")
				Expect(size).To(Equal(newDiskSize))
				Expect(vmCID).To(Equal(newVMCID))
				return newDiskCID, nil
			}
			attachDiskCallCount := 0
			mockCloud.AttachDiskStub = func(vmCID, diskCID string) (interface{}, error) {
				callOrder = append(callOrder, "AttachDisk")
				attachDiskCallCount++
				Expect(vmCID).To(Equal(newVMCID))
				if attachDiskCallCount == 1 {
					Expect(diskCID).To(Equal(oldDiskCID))
					return "/dev/xyz", nil
				}
				Expect(diskCID).To(Equal(newDiskCID))
				return "/dev/abc", nil
			}
		}

		var expectDeployWithDiskMigrationMissingVM = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			callOrder = nil
			wireCommonStubs()
			deleteOldVM1Err = nil

			mockCloud.HasVMStub = func(vmCID string) (bool, error) {
				callOrder = append(callOrder, "HasVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return false, nil
			}
			// delete old vm (without talking to agent) so that the cpi can clean up related resources
			mockCloud.DeleteVMStub = func(vmCID string) error {
				callOrder = append(callOrder, "DeleteVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return deleteOldVM1Err
			}
			mockCloud.CreateVMStub = func(gotAgentID, gotStemcellCID string, cloudProperties biproperty.Map, diskCIDs []string, networks map[string]biproperty.Map, env biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateVM")
				Expect(gotAgentID).To(Equal(agentID))
				Expect(gotStemcellCID).To(Equal(stemcellCID))
				Expect(diskCIDs).To(Equal([]string{oldDiskCID}))
				return newVMCID, nil
			}
			mockCloud.CreateDiskStub = func(size int, cloudProperties biproperty.Map, vmCID string) (string, error) {
				callOrder = append(callOrder, "CreateDisk")
				Expect(size).To(Equal(newDiskSize))
				Expect(vmCID).To(Equal(newVMCID))
				return newDiskCID, nil
			}
			attachDiskCallCount := 0
			mockCloud.AttachDiskStub = func(vmCID, diskCID string) (interface{}, error) {
				callOrder = append(callOrder, "AttachDisk")
				attachDiskCallCount++
				Expect(vmCID).To(Equal(newVMCID))
				if attachDiskCallCount == 1 {
					Expect(diskCID).To(Equal(oldDiskCID))
					return "/dev/xyz", nil
				}
				Expect(diskCID).To(Equal(newDiskCID))
				return "/dev/abc", nil
			}
		}

		var expectDeployWithNoDiskToMigrate = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"

			callOrder = nil
			wireCommonStubs()

			mockCloud.HasVMStub = func(vmCID string) (bool, error) {
				callOrder = append(callOrder, "HasVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return true, nil
			}
			mockCloud.DeleteVMStub = func(vmCID string) error {
				callOrder = append(callOrder, "DeleteVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return nil
			}
			mockAgentClient.ListDiskStub = func() ([]string, error) {
				callOrder = append(callOrder, "ListDisk")
				return []string{oldDiskCID}, nil
			}
			mockCloud.CreateVMStub = func(gotAgentID, gotStemcellCID string, cloudProperties biproperty.Map, diskCIDs []string, networks map[string]biproperty.Map, env biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateVM")
				Expect(gotAgentID).To(Equal(agentID))
				Expect(gotStemcellCID).To(Equal(stemcellCID))
				Expect(diskCIDs).To(Equal([]string{oldDiskCID}))
				return newVMCID, nil
			}
			// attaching a missing disk will fail
			mockCloud.AttachDiskStub = func(vmCID, diskCID string) (interface{}, error) {
				callOrder = append(callOrder, "AttachDisk")
				Expect(vmCID).To(Equal(newVMCID))
				Expect(diskCID).To(Equal(oldDiskCID))
				return "", bicloud.NewCPIError("attach_disk", bicloud.CmdError{
					Type:    bicloud.DiskNotFoundError,
					Message: "fake-disk-not-found-message",
				})
			}
		}

		var expectDeployWithDiskMigrationFailure = func() {
			agentID := "fake-uuid-1"
			oldVMCID := "fake-vm-cid-1"
			newVMCID := "fake-vm-cid-2"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-2"
			newDiskSize := 2048

			callOrder = nil
			wireCommonStubs()

			mockCloud.HasVMStub = func(vmCID string) (bool, error) {
				callOrder = append(callOrder, "HasVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return true, nil
			}
			mockCloud.DeleteVMStub = func(vmCID string) error {
				callOrder = append(callOrder, "DeleteVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return nil
			}
			mockAgentClient.ListDiskStub = func() ([]string, error) {
				callOrder = append(callOrder, "ListDisk")
				return []string{oldDiskCID}, nil
			}
			mockCloud.CreateVMStub = func(gotAgentID, gotStemcellCID string, cloudProperties biproperty.Map, diskCIDs []string, networks map[string]biproperty.Map, env biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateVM")
				Expect(gotAgentID).To(Equal(agentID))
				Expect(gotStemcellCID).To(Equal(stemcellCID))
				Expect(diskCIDs).To(Equal([]string{oldDiskCID}))
				return newVMCID, nil
			}
			mockCloud.CreateDiskStub = func(size int, cloudProperties biproperty.Map, vmCID string) (string, error) {
				callOrder = append(callOrder, "CreateDisk")
				Expect(size).To(Equal(newDiskSize))
				Expect(vmCID).To(Equal(newVMCID))
				return newDiskCID, nil
			}
			attachDiskCallCount := 0
			mockCloud.AttachDiskStub = func(vmCID, diskCID string) (interface{}, error) {
				callOrder = append(callOrder, "AttachDisk")
				attachDiskCallCount++
				Expect(vmCID).To(Equal(newVMCID))
				if attachDiskCallCount == 1 {
					Expect(diskCID).To(Equal(oldDiskCID))
					return "/dev/xyz", nil
				}
				Expect(diskCID).To(Equal(newDiskCID))
				return "/dev/abc", nil
			}
			// migrate fails
			mockAgentClient.MigrateDiskStub = func() error {
				callOrder = append(callOrder, "MigrateDisk")
				return bosherr.Error("fake-migration-error")
			}
		}

		var expectDeployWithDiskMigrationRepair = func(failedMigrationDiskCID string) {
			agentID := "fake-uuid-2"
			oldVMCID := "fake-vm-cid-2"
			newVMCID := "fake-vm-cid-3"
			oldDiskCID := "fake-disk-cid-1"
			newDiskCID := "fake-disk-cid-3"
			newDiskSize := 2048

			callOrder = nil
			wireCommonStubs()

			mockCloud.HasVMStub = func(vmCID string) (bool, error) {
				callOrder = append(callOrder, "HasVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return true, nil
			}
			mockCloud.DeleteVMStub = func(vmCID string) error {
				callOrder = append(callOrder, "DeleteVM")
				Expect(vmCID).To(Equal(oldVMCID))
				return nil
			}
			mockAgentClient.ListDiskStub = func() ([]string, error) {
				callOrder = append(callOrder, "ListDisk")
				return []string{oldDiskCID}, nil
			}
			mockCloud.CreateVMStub = func(gotAgentID, gotStemcellCID string, cloudProperties biproperty.Map, diskCIDs []string, networks map[string]biproperty.Map, env biproperty.Map) (string, error) {
				callOrder = append(callOrder, "CreateVM")
				Expect(gotAgentID).To(Equal(agentID))
				Expect(gotStemcellCID).To(Equal(stemcellCID))
				Expect(diskCIDs).To(Equal([]string{oldDiskCID, failedMigrationDiskCID}))
				return newVMCID, nil
			}
			mockCloud.CreateDiskStub = func(size int, cloudProperties biproperty.Map, vmCID string) (string, error) {
				callOrder = append(callOrder, "CreateDisk")
				Expect(size).To(Equal(newDiskSize))
				Expect(vmCID).To(Equal(newVMCID))
				return newDiskCID, nil
			}
			attachDiskCallCount := 0
			mockCloud.AttachDiskStub = func(vmCID, diskCID string) (interface{}, error) {
				callOrder = append(callOrder, "AttachDisk")
				attachDiskCallCount++
				Expect(vmCID).To(Equal(newVMCID))
				if attachDiskCallCount == 1 {
					Expect(diskCID).To(Equal(oldDiskCID))
					return "/dev/xyz", nil
				}
				Expect(diskCID).To(Equal(newDiskCID))
				return "/dev/abc", nil
			}
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

			fakeDigestCalculator = &cryptofakes.FakeDigestCalculator{}

			mockInstaller = &installationfakes.FakeInstaller{}
			mockInstallerFactory = &installationfakes.FakeInstallerFactory{}
			mockCloudFactory = &cloudfakes.FakeFactory{}

			sshTunnelFactory = bisshtunnel.NewFactory(logger)

			fakeRepoUUIDGenerator = fakeuuid.NewFakeGenerator()

			mockCloud = &cloudfakes.FakeCloud{}

			releaseReader = &fakerel.FakeReader{}
			releaseManager = biinstall.NewReleaseManager(logger)

			mockStateBuilderFactory = &statefakes.FakeBuilderFactory{}
			mockStateBuilder = &statefakes.FakeBuilder{}
			mockState = &statefakes.FakeState{}

			mockBlobstoreFactory = &blobstorefakes.FakeFactory{}
			mockBlobstore = &blobstorefakes.FakeBlobstore{}
			mockBlobstoreFactory.CreateReturns(mockBlobstore, nil)

			fakeStemcellExtractor = fakebistemcell.NewFakeExtractor()

			stdOut = gbytes.NewBuffer()
			stdErr = gbytes.NewBuffer()
			fakeStage = &testui.Stage{}

			mockAgentClientFactory = &fakecmd.FakeAgentClientFactory{}
			mockAgentClient = &agentclientfakes.FakeAgentClient{}

			mockAgentClientFactory.NewAgentClientReturns(mockAgentClient, nil)

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

			Expect(callOrder).To(Equal([]string{
				"CreateStemcell", "CreateVM", "SetVMMetadata", "Ping",
				"CreateDisk", "AttachDisk", "SetDiskMetadata", "Ping", "AddPersistentDisk", "MountDisk",
				"Apply", "GetState", "Stop", "Apply", "RunScript:pre-start", "Start", "GetState", "RunScript:post-start",
			}))
		})

		Context("when multiple releases are provided", func() {
			var (
				otherReleaseTarballPath = filepath.Join("/", "fake-other-release.tgz")
			)

			BeforeEach(func() {
				err := fs.WriteFileString(otherReleaseTarballPath, "fake-other-tgz-content")
				Expect(err).ToNot(HaveOccurred())

				job := bireljob.NewJob(resource.NewResource("other", "", nil))

				otherRelease := birel.NewRelease(
					"fake-other-release-name",
					"1.2",
					"commit",
					false,
					[]*bireljob.Job{job},
					[]*birelpkg.Package{},
					nil,
					nil,
					false,
					"fake-other-extracted-dir",
					fs,
				)
				releaseReader.ReadStub = func(path string) (birel.Release, error) {
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

				// a new directorID will be generated; mockAgentClientFactory.NewAgentClientReturns
				// (set in the outer BeforeEach) already covers whatever directorID is passed in.

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

					Expect(callOrder).To(Equal([]string{
						"HasVM", "Ping", "RunScript:pre-stop", "Drain", "Stop", "RunScript:post-stop", "ListDisk", "UnmountDisk", "DeleteVM",
						"CreateVM", "SetVMMetadata", "Ping",
						"AttachDisk", "SetDiskMetadata", "Ping", "AddPersistentDisk", "MountDisk",
						"CreateDisk", "AttachDisk", "SetDiskMetadata", "Ping", "AddPersistentDisk", "MountDisk",
						"MigrateDisk", "RemovePersistentDisk", "DetachDisk", "Ping", "DeleteDisk",
						"Apply", "GetState", "Stop", "Apply", "RunScript:pre-start", "Start", "GetState", "RunScript:post-start",
					}))
				})

				Context("when current VM has been deleted manually (outside of bosh)", func() {
					It("migrates the disk content, but does not shutdown the old VM", func() {
						expectDeployWithDiskMigrationMissingVM()

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
						Expect(err).ToNot(HaveOccurred())

						Expect(callOrder).To(Equal([]string{
							"HasVM", "DeleteVM",
							"CreateVM", "SetVMMetadata", "Ping",
							"AttachDisk", "SetDiskMetadata", "Ping", "AddPersistentDisk", "MountDisk",
							"CreateDisk", "AttachDisk", "SetDiskMetadata", "Ping", "AddPersistentDisk", "MountDisk",
							"MigrateDisk", "RemovePersistentDisk", "DetachDisk", "Ping", "DeleteDisk",
							"Apply", "GetState", "Stop", "Apply", "RunScript:pre-start", "Start", "GetState", "RunScript:post-start",
						}))
					})

					It("ignores DiskNotFound errors", func() {
						expectDeployWithDiskMigrationMissingVM()

						deleteOldVM1Err = bicloud.NewCPIError("delete_vm", bicloud.CmdError{
							Type:    bicloud.VMNotFoundError,
							Message: "fake-vm-not-found-message",
						})

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
						failedMigrationDiskCID := "fake-disk-cid-2"
						expectDeployWithDiskMigrationRepair(failedMigrationDiskCID)

						err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
						Expect(err).ToNot(HaveOccurred())

						deletedDiskCIDs := []string{}
						for i := 0; i < mockCloud.DeleteDiskCallCount(); i++ {
							deletedDiskCIDs = append(deletedDiskCIDs, mockCloud.DeleteDiskArgsForCall(i))
						}
						Expect(deletedDiskCIDs).To(ContainElement(failedMigrationDiskCID))

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
				mockCloud.HasVMReturns(true, nil)
			}

			Context("and the same deployment is attempted again", func() {
				It("skips the deploy", func() {
					deleteVMCountBefore := mockCloud.DeleteVMCallCount()
					createVMCountBefore := mockCloud.CreateVMCallCount()

					expectNoDeployHappened()

					err := newCreateEnvCmd().Run(fakeStage, newDeployOpts(deploymentManifestPath, ""))
					Expect(err).ToNot(HaveOccurred())
					Expect(stdOut).To(gbytes.Say("No deployment, stemcell or release changes. Skipping deploy."))

					Expect(mockCloud.DeleteVMCallCount()).To(Equal(deleteVMCountBefore))
					Expect(mockCloud.CreateVMCallCount()).To(Equal(createVMCountBefore))
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

				// a new directorID will be generated; mockCloudFactory.NewCloudReturns and
				// mockAgentClientFactory.NewAgentClientReturns already cover any directorID passed in.

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

func newDeployOpts(manifestPath string, statePath string) opts.CreateEnvOpts {
	return opts.CreateEnvOpts{StatePath: statePath, Args: opts.CreateEnvArgs{Manifest: opts.FileBytesWithPathArg{Path: manifestPath}}}
}
