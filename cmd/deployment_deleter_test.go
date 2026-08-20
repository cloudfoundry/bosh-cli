package cmd_test

import (
	"errors"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/agentclient/agentclientfakes"
	"github.com/cloudfoundry/bosh-cli/v7/blobstore/blobstorefakes"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bicpirel "github.com/cloudfoundry/bosh-cli/v7/cpi/release"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/deploymentfakes"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	"github.com/cloudfoundry/bosh-cli/v7/installation/installationfakes"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
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
	Describe("DeleteDeployment", func() {
		var (
			fs                          *fakesys.FakeFileSystem
			logger                      boshlog.Logger
			releaseReader               *fakerel.FakeReader
			releaseManager              boshrel.Manager
			mockCpiInstaller            *installationfakes.FakeInstaller
			mockCpiUninstaller          *installationfakes.FakeUninstaller
			mockInstallerFactory        *installationfakes.FakeInstallerFactory
			mockCloudFactory            *cloudfakes.FakeFactory
			fakeUUIDGenerator           *fakeuuid.FakeGenerator
			setupDeploymentStateService biconfig.DeploymentStateService
			fakeInstallation            *fakecmd.FakeInstallation

			fakeUI *fakeui.FakeUI

			mockBlobstoreFactory *blobstorefakes.FakeFactory
			mockBlobstore        *blobstorefakes.FakeBlobstore

			mockDeploymentManagerFactory *deploymentfakes.FakeManagerFactory
			mockDeploymentManager        *deploymentfakes.FakeManager
			mockDeployment               *deploymentfakes.FakeDeployment

			mockAgentClient        *agentclientfakes.FakeAgentClient
			mockAgentClientFactory *fakecmd.FakeAgentClientFactory
			mockCloud              *cloudfakes.FakeCloud

			fakeStage *fakeui.FakeStage

			directorID string

			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			deploymentStatePath    string

			// callOrder records, in invocation order, calls to Install/NewCloud/Delete/Cleanup
			// made during a test -- the counterfeiter equivalent of gomock.InOrder().
			callOrder []string

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
				false,
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
			mockInstallerFactory.NewInstallerReturns(mockCpiInstaller)

			mockCpiInstaller.InstallStub = func(manifest biinstallmanifest.Manifest, stage boshui.Stage) (biinstall.Installation, error) {
				callOrder = append(callOrder, "Install")
				Expect(fakeStage.SubStages).To(ContainElement(stage))
				return fakeInstallation, nil
			}
			mockCpiInstaller.CleanupReturns(nil)

			mockCloudFactory.NewCloudStub = func(installation biinstall.Installation, directorID string, stemcellApiVersion int) (bicloud.Cloud, error) {
				callOrder = append(callOrder, "NewCloud")
				return mockCloud, nil
			}
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
				"",
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
			mockDeploymentManagerFactory.NewManagerReturns(mockDeploymentManager)
			mockDeploymentManager.FindCurrentReturns(mockDeployment, true, nil)

			mockDeployment.DeleteStub = func(_ bool, stage boshui.Stage) error {
				callOrder = append(callOrder, "Delete")
				Expect(fakeStage.SubStages).To(ContainElement(stage))
				return nil
			}
			mockDeploymentManager.CleanupStub = func(stage boshui.Stage) error {
				callOrder = append(callOrder, "Cleanup")
				return nil
			}
			if defaultUninstallerUsed {
				mockCpiUninstaller.UninstallReturns(nil)
			}
		}

		var expectCleanup = func() {
			mockDeploymentManagerFactory.NewManagerReturns(mockDeploymentManager)
			mockDeploymentManager.FindCurrentReturns(nil, false, nil)

			mockDeploymentManager.CleanupReturns(nil)
			mockCpiUninstaller.UninstallReturns(nil)
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

			callOrder = nil

			mockCloud = &cloudfakes.FakeCloud{}
			mockCloudFactory = &cloudfakes.FakeFactory{}

			mockCpiInstaller = &installationfakes.FakeInstaller{}
			mockCpiUninstaller = &installationfakes.FakeUninstaller{}
			mockInstallerFactory = &installationfakes.FakeInstallerFactory{}

			fakeInstallation = &fakecmd.FakeInstallation{}

			mockBlobstoreFactory = &blobstorefakes.FakeFactory{}
			mockBlobstore = &blobstorefakes.FakeBlobstore{}
			mockBlobstoreFactory.CreateReturns(mockBlobstore, nil)

			mockDeploymentManagerFactory = &deploymentfakes.FakeManagerFactory{}
			mockDeploymentManager = &deploymentfakes.FakeManager{}
			mockDeployment = &deploymentfakes.FakeDeployment{}

			releaseReader = &fakerel.FakeReader{}
			releaseManager = biinstall.NewReleaseManager(logger)

			mockAgentClientFactory = &fakecmd.FakeAgentClientFactory{}
			mockAgentClient = &agentclientfakes.FakeAgentClient{}

			directorID = "fake-uuid-0"
			skipDrain = false

			mockAgentClientFactory.NewAgentClientReturns(mockAgentClient, nil)

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

						Expect(mockCloudFactory.NewCloudCallCount()).To(Equal(1))
						_, gotDirectorID, gotStemcellApiVersion := mockCloudFactory.NewCloudArgsForCall(0)
						Expect(gotDirectorID).To(Equal(directorID))
						Expect(gotStemcellApiVersion).To(Equal(stemcellApiVersionForDelete))
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

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)
					Expect(err).NotTo(HaveOccurred())

					installIndex := -1
					newCloudIndex := -1
					for i, name := range callOrder {
						switch name {
						case "Install":
							installIndex = i
						case "NewCloud":
							newCloudIndex = i
						}
					}
					Expect(installIndex).To(BeNumerically(">=", 0), "Install should have been called")
					Expect(newCloudIndex).To(BeNumerically(">", installIndex), "NewCloud should have been called after Install")
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

					Expect(mockBlobstoreFactory.CreateCallCount()).To(Equal(1))
					gotURL, gotClient := mockBlobstoreFactory.CreateArgsForCall(0)
					Expect(gotURL).To(Equal(mbusURL))
					Expect(gotClient).To(SecureTLSClientMatcher())

					Expect(mockAgentClientFactory.NewAgentClientCallCount()).To(Equal(1))
					gotDirectorID, gotMbusURL, gotCert := mockAgentClientFactory.NewAgentClientArgsForCall(0)
					Expect(gotDirectorID).To(Equal(directorID))
					Expect(gotMbusURL).To(Equal(mbusURL))
					Expect(gotCert).To(Equal(certificate))
				})

				It("deletes the local CPI installation", func() {
					expectDeleteAndCleanup(skipDrain, false)
					mockCpiUninstaller.UninstallReturns(nil)

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
				mockInstallerFactory.NewInstallerReturns(mockCpiInstaller)

				fakeInstallation := &fakecmd.FakeInstallation{}

				mockCpiInstaller.InstallStub = func(manifest biinstallmanifest.Manifest, stage boshui.Stage) (biinstall.Installation, error) {
					Expect(fakeStage.SubStages).To(ContainElement(stage))
					return fakeInstallation, nil
				}
				mockCpiInstaller.CleanupReturns(nil)

				mockCloudFactory.NewCloudReturns(mockCloud, nil)
			})

			Context("when the call to delete the deployment returns an error", func() {
				It("returns the error", func() {
					mockDeploymentManagerFactory.NewManagerReturns(mockDeploymentManager)
					mockDeploymentManager.FindCurrentReturns(mockDeployment, true, nil)

					deleteError := bosherr.Error("delete error")

					mockDeployment.DeleteReturns(deleteError)

					err := newDeploymentDeleter().DeleteDeployment(skipDrain, fakeStage)

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
