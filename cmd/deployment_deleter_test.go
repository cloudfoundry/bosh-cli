package cmd_test

import (
	"errors"
	"os"
	"path/filepath"

	mockhttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http/mocks"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/cppforlife/go-patch/patch"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mockagentclient "github.com/cloudfoundry/bosh-cli/v7/agentclient/mocks"
	mockblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore/mocks"
	mockcloud "github.com/cloudfoundry/bosh-cli/v7/cloud/mocks"
	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	mockdeployment "github.com/cloudfoundry/bosh-cli/v7/deployment/mocks"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	mockinstall "github.com/cloudfoundry/bosh-cli/v7/installation/mocks"
	bitarball "github.com/cloudfoundry/bosh-cli/v7/installation/tarball"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshjob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeploymentDeleter", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("DeleteDeployment", func() {
		var (
			fs                          *fakesys.FakeFileSystem
			logger                      boshlog.Logger
			releaseReader               *fakerel.FakeReader
			releaseManager              boshrel.Manager
			mockCpiInstaller            *mockinstall.MockInstaller
			mockCpiUninstaller          *mockinstall.MockUninstaller
			mockInstallerFactory        *mockinstall.MockInstallerFactory
			mockCloudFactory            *mockcloud.MockFactory
			fakeUUIDGenerator           *fakeuuid.FakeGenerator
			setupDeploymentStateService biconfig.DeploymentStateService
			fakeInstallation            *fakecmd.FakeInstallation

			fakeUI *fakeui.FakeUI

			mockBlobstoreFactory *mockblobstore.MockFactory
			mockBlobstore        *mockblobstore.MockBlobstore

			mockDeploymentManagerFactory *mockdeployment.MockManagerFactory
			mockDeploymentManager        *mockdeployment.MockManager
			mockDeployment               *mockdeployment.MockDeployment

			mockAgentClient        *mockagentclient.MockAgentClient
			mockAgentClientFactory *mockhttpagent.MockAgentClientFactory
			mockCloud              *mockcloud.MockCloud

			fakeStage *fakeui.FakeStage

			directorID string

			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			deploymentStatePath    string

			expectCPIInstall *gomock.Call
			expectNewCloud   *gomock.Call

			mbusURL                     = "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint"
			stemcellApiVersionForDelete = 1
			skipDrain                   bool
		)

		var certificate = `-----BEGIN CERTIFICATE-----
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

		var writeDeploymentManifest = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---
name: test-release

releases:
- name: fake-cpi-release-name
  url: file:///fake-cpi-release.tgz

cloud_provider:
  template:
    name: fake-cpi-release-job-name
    release: fake-cpi-release-name
  mbus: http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint
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
`)
			Expect(err).ToNot(HaveOccurred())
		}

		var writeCPIReleaseTarball = func() {
			err := fs.WriteFileString("/fake-cpi-release.tgz", "fake-tgz-content")
			Expect(err).ToNot(HaveOccurred())
		}

		var allowCPIToBeExtracted = func() {
			job := boshjob.NewJob(NewResource("fake-cpi-release-job-name", "job-fp", nil))
			job.Templates = map[string]string{"templates/cpi.erb": "bin/cpi"}

			cpiRelease := boshrel.NewRelease(
				"fake-cpi-release-name",
				"fake-cpi-release-version",
				"fake-sha",
				false,
				[]*boshjob.Job{job},
				[]*boshpkg.Package{},
				[]*boshpkg.CompiledPackage{},
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
		}

		var allowCPIToBeInstalled = func() {
			installationManifest := biinstallmanifest.Manifest{
				Name: "test-release",
				Template: biinstallmanifest.ReleaseJobRef{
					Name:    "fake-cpi-release-job-name",
					Release: "fake-cpi-release-name",
				},
				Mbus:       mbusURL,
				Properties: biproperty.Map{},
				Cert: biinstallmanifest.Certificate{
					CA: certificate,
				},
			}

			target := biinstall.NewTarget(filepath.Join("fake-install-dir", "fake-installation-id"))
			mockInstallerFactory.EXPECT().NewInstaller(target).Return(mockCpiInstaller).AnyTimes()

			expectCPIInstall = mockCpiInstaller.EXPECT().Install(installationManifest, gomock.Any()).Do(func(_ biinstallmanifest.Manifest, stage boshui.Stage) {
				Expect(fakeStage.SubStages).To(ContainElement(stage))
			}).Return(fakeInstallation, nil).AnyTimes()
			mockCpiInstaller.EXPECT().Cleanup(fakeInstallation).AnyTimes()

			expectNewCloud = mockCloudFactory.EXPECT().NewCloud(fakeInstallation, directorID, stemcellApiVersionForDelete).Return(mockCloud, nil).AnyTimes()
		}

		var newDeploymentDeleter = func() cmd.DeploymentDeleter {
			releaseSetValidator := birelsetmanifest.NewValidator(logger)
			releaseSetParser := birelsetmanifest.NewParser(fs, logger, releaseSetValidator)
			installationValidator := biinstallmanifest.NewValidator(logger)
			installationParser := biinstallmanifest.NewParser(fs, fakeUUIDGenerator, logger, installationValidator)
			tarballCache := bitarball.NewCache("fake-base-path", fs, logger)
			tarballProvider := bitarball.NewProvider(tarballCache, fs, nil, 1, 0, logger)
			deploymentStateService := biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath, ""))

			cpiInstaller := bicpirel.CpiInstaller{
				ReleaseManager:   releaseManager,
				InstallerFactory: mockInstallerFactory,
				Validator:        bicpirel.NewValidator(),
			}
			releaseFetcher := biinstall.NewReleaseFetcher(tarballProvider, releaseReader, releaseManager)
			releaseSetAndInstallationManifestParser := cmd.ReleaseSetAndInstallationManifestParser{
				ReleaseSetParser:   releaseSetParser,
				InstallationParser: installationParser,
			}
			fakeInstallationUUIDGenerator := &fakeuuid.FakeGenerator{}
			fakeInstallationUUIDGenerator.GeneratedUUID = "fake-installation-id"
			targetProvider := biinstall.NewTargetProvider(
				deploymentStateService,
				fakeInstallationUUIDGenerator,
				filepath.Join("fake-install-dir"),
			)

			tempRootConfigurator := cmd.NewTempRootConfigurator(fs)

			return cmd.NewDeploymentDeleter(
				fakeUI,
				"deleteCmd",
				logger,
				deploymentStateService,
				releaseManager,
				mockCloudFactory,
				mockAgentClientFactory,
				mockBlobstoreFactory,
				mockDeploymentManagerFactory,
				deploymentManifestPath,
				boshtpl.StaticVariables{},
				patch.Ops{},
				cpiInstaller,
				mockCpiUninstaller,
				releaseFetcher,
				releaseSetAndInstallationManifestParser,
				tempRootConfigurator,
				targetProvider,
			)
		}

		var expectDeleteAndCleanup = func(skipDrain, defaultUninstallerUsed bool) {
			mockDeploymentManagerFactory.EXPECT().NewManager(mockCloud, mockAgentClient, mockBlobstore).Return(mockDeploymentManager)
			mockDeploymentManager.EXPECT().FindCurrent().Return(mockDeployment, true, nil)

			gomock.InOrder(
				mockDeployment.EXPECT().Delete(skipDrain, gomock.Any()).Do(func(_ bool, stage boshui.Stage) {
					Expect(fakeStage.SubStages).To(ContainElement(stage))
				}),
				mockDeploymentManager.EXPECT().Cleanup(fakeStage),
			)
			if defaultUninstallerUsed {
				mockCpiUninstaller.EXPECT().Uninstall(gomock.Any()).Return(nil)
			}
		}

		var expectCleanup = func() {
			mockDeploymentManagerFactory.EXPECT().NewManager(mockCloud, mockAgentClient, mockBlobstore).Return(mockDeploymentManager).AnyTimes()
			mockDeploymentManager.EXPECT().FindCurrent().Return(nil, false, nil).AnyTimes()

			mockDeploymentManager.EXPECT().Cleanup(fakeStage)
			mockCpiUninstaller.EXPECT().Uninstall(gomock.Any()).Return(nil)
		}

		var expectValidationInstallationDeletionEvents = func() {
			Expect(fakeUI.Said).To(Equal([]string{
				"Deployment state: '" + filepath.Join("/", "deployment-dir", "fake-deployment-manifest-state.json") + "'\n",
			}))

			Expect(fakeStage.PerformCalls).To(Equal([]*fakeui.PerformCall{
				{
					Name: "validating",
					Stage: &fakeui.FakeStage{
						PerformCalls: []*fakeui.PerformCall{
							{Name: "Validating release 'fake-cpi-release-name'"},
							{Name: "Validating cpi release"},
						},
					},
				},
				{
					Name:  "installing CPI",
					Stage: &fakeui.FakeStage{},
				},
				{
					Name:  "deleting deployment",
					Stage: &fakeui.FakeStage{},
				},
				{
					Name: "Uninstalling local artifacts for CPI and deployment",
				},
				{
					Name: "Cleaning up rendered CPI jobs",
				},
				// mock deployment manager cleanup doesn't add sub-stages
			}))

			// installing steps handled by installer.Install()
			// deleting steps handled by deployment.Delete()
		}

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
			deploymentStatePath = biconfig.DeploymentStatePath(deploymentManifestPath, "")
			setupDeploymentStateService = biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, deploymentStatePath)
			_, err := setupDeploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())

			fakeUI = &fakeui.FakeUI{}

			fakeStage = fakeui.NewFakeStage()

			mockCloud = mockcloud.NewMockCloud(mockCtrl)
			mockCloudFactory = mockcloud.NewMockFactory(mockCtrl)

			mockCpiInstaller = mockinstall.NewMockInstaller(mockCtrl)
			mockCpiUninstaller = mockinstall.NewMockUninstaller(mockCtrl)
			mockInstallerFactory = mockinstall.NewMockInstallerFactory(mockCtrl)

			fakeInstallation = &fakecmd.FakeInstallation{}

			mockBlobstoreFactory = mockblobstore.NewMockFactory(mockCtrl)
			mockBlobstore = mockblobstore.NewMockBlobstore(mockCtrl)
			mockBlobstoreFactory.EXPECT().Create(mbusURL, gomock.Any()).Return(mockBlobstore, nil).AnyTimes()

			mockDeploymentManagerFactory = mockdeployment.NewMockManagerFactory(mockCtrl)
			mockDeploymentManager = mockdeployment.NewMockManager(mockCtrl)
			mockDeployment = mockdeployment.NewMockDeployment(mockCtrl)

			releaseReader = &fakerel.FakeReader{}
			releaseManager = biinstall.NewReleaseManager(logger)

			mockAgentClientFactory = mockhttpagent.NewMockAgentClientFactory(mockCtrl)
			mockAgentClient = mockagentclient.NewMockAgentClient(mockCtrl)

			directorID = "fake-uuid-0"
			skipDrain = false

			mockAgentClientFactory.EXPECT().NewAgentClient(
				directorID,
				"http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint",
				certificate,
			).Return(mockAgentClient, nil).AnyTimes()

			writeDeploymentManifest()
			writeCPIReleaseTarball()

			stemcellApiVersionForDelete = 1
		})

		JustBeforeEach(func() {
			allowCPIToBeExtracted()
		})

		Context("when the CPI installs", func() {

			JustBeforeEach(func() {
				allowCPIToBeInstalled()
			})

			Context("when the deployment state file does not exist", func() {
				BeforeEach(func() {
					err := fs.RemoveAll(deploymentStatePath)
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete anything", func() {
					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeUI.Said).To(Equal([]string{
						"Deployment state: '" + filepath.Join("/", "deployment-dir", "fake-deployment-manifest-state.json") + "'\n",
						"No deployment state file found.\n",
					}))
				})
			})

			Context("when the deployment has been deployed", func() {
				BeforeEach(func() {
					// create deployment manifest yaml file
					err := setupDeploymentStateService.Save(biconfig.DeploymentState{
						DirectorID: directorID,
					})
					Expect(err).ToNot(HaveOccurred())
				})

				Context("stemcell version is 2 and present in deployment state", func() {
					BeforeEach(func() {
						err := setupDeploymentStateService.Save(biconfig.DeploymentState{
							DirectorID:        directorID,
							CurrentStemcellID: "stemcell-id",
							Stemcells: []biconfig.StemcellRecord{
								{
									ID:         "stemcell-id",
									ApiVersion: 2,
								},
							},
						})
						Expect(err).ToNot(HaveOccurred())

						stemcellApiVersionForDelete = 2
					})

					It("sets stemcell version for cloud", func() {
						expectDeleteAndCleanup(true, true)
						err := newDeploymentDeleter().DeleteDeployment(true, fakeStage)
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("when change temp root fails", func() {
					It("returns an error", func() {
						fs.ChangeTempRootErr = errors.New("fake ChangeTempRootErr")
						err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("Setting temp root: fake ChangeTempRootErr"))
					})
				})

				It("sets the temp root", func() {
					expectDeleteAndCleanup(skipDrain, true)
					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).NotTo(HaveOccurred())
					Expect(fs.TempRootPath).To(Equal(filepath.Join("fake-install-dir", "fake-installation-id", "tmp")))
				})

				It("extracts & install CPI release tarball", func() {
					expectDeleteAndCleanup(skipDrain, true)

					gomock.InOrder(
						expectCPIInstall.Times(1),
						expectNewCloud.Times(1),
					)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).NotTo(HaveOccurred())
				})

				It("deletes the extracted CPI release", func() {
					expectDeleteAndCleanup(skipDrain, true)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).NotTo(HaveOccurred())
					Expect(fs.FileExists("fake-cpi-extracted-dir")).To(BeFalse())
				})

				It("deletes the deployment & cleans up orphans", func() {
					expectDeleteAndCleanup(skipDrain, true)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeUI.Errors).To(BeEmpty())
				})

				It("deletes the local CPI installation", func() {
					expectDeleteAndCleanup(skipDrain, false)
					mockCpiUninstaller.EXPECT().Uninstall(gomock.Any()).Return(nil)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})

				It("logs validating & deleting stages", func() {
					expectDeleteAndCleanup(skipDrain, true)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					expectValidationInstallationDeletionEvents()
				})

				It("deletes the local deployment state file", func() {
					expectDeleteAndCleanup(skipDrain, true)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fs.FileExists(deploymentStatePath)).To(BeFalse())
				})

				It("skips draining if specified", func() {
					skipDrain = true
					expectDeleteAndCleanup(skipDrain, true)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("when nothing has been deployed", func() {
				BeforeEach(func() {
					err := setupDeploymentStateService.Save(biconfig.DeploymentState{DirectorID: "fake-uuid-0"})
					Expect(err).ToNot(HaveOccurred())
				})

				It("cleans up orphans, but does not delete any deployment", func() {
					expectCleanup()

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeUI.Errors).To(BeEmpty())
				})
			})
		})

		Context("when the CPI fails to Delete", func() {
			JustBeforeEach(func() {
				installationManifest := biinstallmanifest.Manifest{
					Name: "test-release",
					Template: biinstallmanifest.ReleaseJobRef{
						Name:    "fake-cpi-release-job-name",
						Release: "fake-cpi-release-name",
					},
					Mbus:       mbusURL,
					Properties: biproperty.Map{},
					Cert: biinstallmanifest.Certificate{
						CA: certificate,
					},
				}

				target := biinstall.NewTarget(filepath.Join("fake-install-dir", "fake-installation-id"))
				mockInstallerFactory.EXPECT().NewInstaller(target).Return(mockCpiInstaller).AnyTimes()

				fakeInstallation := &fakecmd.FakeInstallation{}

				expectCPIInstall = mockCpiInstaller.EXPECT().Install(installationManifest, gomock.Any()).Do(func(_ biinstallmanifest.Manifest, stage boshui.Stage) {
					Expect(fakeStage.SubStages).To(ContainElement(stage))
				}).Return(fakeInstallation, nil).AnyTimes()
				mockCpiInstaller.EXPECT().Cleanup(fakeInstallation).AnyTimes()

				expectNewCloud = mockCloudFactory.EXPECT().NewCloud(fakeInstallation, directorID, stemcellApiVersionForDelete).Return(mockCloud, nil).AnyTimes()
			})

			Context("when the call to delete the deployment returns an error", func() {
				It("returns the error", func() {
					mockDeploymentManagerFactory.EXPECT().NewManager(mockCloud, mockAgentClient, mockBlobstore).Return(mockDeploymentManager)
					mockDeploymentManager.EXPECT().FindCurrent().Return(mockDeployment, true, nil)

					deleteError := bosherr.Error("delete error")

					mockDeployment.EXPECT().Delete(skipDrain, gomock.Any()).Return(deleteError)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
