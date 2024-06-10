package cmd_test

import (
	"path/filepath"

	mockhttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http/mocks"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/cppforlife/go-patch/patch"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mockagentclient "github.com/cloudfoundry/bosh-cli/v7/agentclient/mocks"
	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	mockdeployment "github.com/cloudfoundry/bosh-cli/v7/deployment/mocks"
	bidepltpl "github.com/cloudfoundry/bosh-cli/v7/deployment/template"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	biinstall "github.com/cloudfoundry/bosh-cli/v7/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/v7/release/set/manifest"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeploymentStateManager", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		fs                          *fakesys.FakeFileSystem
		logger                      boshlog.Logger
		fakeUUIDGenerator           *fakeuuid.FakeGenerator
		setupDeploymentStateService biconfig.DeploymentStateService

		fakeUI *fakeui.FakeUI

		mockDeploymentManagerFactory *mockdeployment.MockManagerFactory
		mockDeploymentManager        *mockdeployment.MockManager
		mockDeployment               *mockdeployment.MockDeployment

		mockAgentClient        *mockagentclient.MockAgentClient
		mockAgentClientFactory *mockhttpagent.MockAgentClientFactory

		fakeStage *fakeui.FakeStage

		directorID string

		deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
		deploymentStatePath    string

		skipDrain bool
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

	var newDeploymentStateManager = func() cmd.DeploymentStateManager {
		releaseSetValidator := birelsetmanifest.NewValidator(logger)
		releaseSetParser := birelsetmanifest.NewParser(fs, logger, releaseSetValidator)
		installationValidator := biinstallmanifest.NewValidator(logger)
		installationParser := biinstallmanifest.NewParser(fs, fakeUUIDGenerator, logger, installationValidator)
		deploymentStateService := biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, biconfig.DeploymentStatePath(deploymentManifestPath, ""))

		releaseSetAndInstallationManifestParser := cmd.ReleaseSetAndInstallationManifestParser{
			ReleaseSetParser:   releaseSetParser,
			InstallationParser: installationParser,
		}
		deploymentParser := bideplmanifest.NewParser(fs, logger)
		deploymentValidator := bideplmanifest.NewValidator(logger)
		releaseManager := biinstall.NewReleaseManager(logger)
		deploymentManifestParser := cmd.NewDeploymentManifestParser(
			deploymentParser,
			deploymentValidator,
			releaseManager,
			bidepltpl.NewDeploymentTemplateFactory(fs),
		)

		return cmd.NewDeploymentStateManager(
			fakeUI,
			"deleteCmd",
			logger,
			deploymentStateService,
			mockAgentClientFactory,
			mockDeploymentManagerFactory,
			deploymentManifestPath,
			boshtpl.StaticVariables{},
			patch.Ops{},
			releaseSetAndInstallationManifestParser,
			deploymentManifestParser,
		)
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

		mockDeploymentManagerFactory = mockdeployment.NewMockManagerFactory(mockCtrl)
		mockDeploymentManager = mockdeployment.NewMockManager(mockCtrl)
		mockDeployment = mockdeployment.NewMockDeployment(mockCtrl)

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
	})

	Describe("StopDeployment", func() {
		var expectStop = func(skipDrain bool) {
			mockDeploymentManagerFactory.EXPECT().NewManager(gomock.Any(), mockAgentClient, gomock.Any()).AnyTimes().Return(mockDeploymentManager)
			mockDeploymentManager.EXPECT().FindCurrent().Return(mockDeployment, true, nil)

			gomock.InOrder(
				mockDeployment.EXPECT().Stop(skipDrain, gomock.Any()).Do(func(_ bool, stage boshui.Stage) {
					Expect(fakeStage.SubStages).To(ContainElement(stage))
				}),
			)
		}

		var expectValidationStopEvents = func() {
			Expect(fakeUI.Said).To(Equal([]string{
				"Deployment state: '" + filepath.Join("/", "deployment-dir", "fake-deployment-manifest-state.json") + "'\n",
			}))

			Expect(fakeStage.PerformCalls).To(Equal([]*fakeui.PerformCall{
				{
					Name: "validating",
					Stage: &fakeui.FakeStage{
						PerformCalls: []*fakeui.PerformCall{
							{Name: "Validating deployment manifest"},
						},
					},
				},
				{
					Name:  "stopping deployment",
					Stage: &fakeui.FakeStage{},
				},
			}))
		}

		Context("when the deployment state file does not exist", func() {
			BeforeEach(func() {
				err := fs.RemoveAll(deploymentStatePath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not stop anything", func() {
				err := newDeploymentStateManager().StopDeployment(skipDrain, fakeStage)
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

			It("stops the deployment", func() {
				expectStop(skipDrain)

				err := newDeploymentStateManager().StopDeployment(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUI.Errors).To(BeEmpty())
			})

			It("logs validating & stop stages", func() {
				expectStop(skipDrain)

				err := newDeploymentStateManager().StopDeployment(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				expectValidationStopEvents()
			})

			It("skips draining if specified", func() {
				skipDrain = true
				expectStop(skipDrain)

				err := newDeploymentStateManager().StopDeployment(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when nothing has been deployed", func() {
			BeforeEach(func() {
				err := setupDeploymentStateService.Save(biconfig.DeploymentState{DirectorID: "fake-uuid-0"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("tries to stop deployment", func() {
				expectStop(skipDrain)

				err := newDeploymentStateManager().StopDeployment(skipDrain, fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUI.Errors).To(BeEmpty())
			})
		})
	})

	Describe("StartDeployment", func() {

		var expectStart = func() {
			mockDeploymentManagerFactory.EXPECT().NewManager(gomock.Any(), mockAgentClient, gomock.Any()).AnyTimes().Return(mockDeploymentManager)
			mockDeploymentManager.EXPECT().FindCurrent().Return(mockDeployment, true, nil)

			gomock.InOrder(
				mockDeployment.EXPECT().Start(gomock.Any(), gomock.Any()).Do(func(stage boshui.Stage, update bideplmanifest.Update) {
					Expect(fakeStage.SubStages).To(ContainElement(stage))
					Expect(update).ToNot(BeNil())
				}),
			)
		}

		var expectValidationStartEvents = func() {
			Expect(fakeUI.Said).To(Equal([]string{
				"Deployment state: '" + filepath.Join("/", "deployment-dir", "fake-deployment-manifest-state.json") + "'\n",
			}))

			Expect(fakeStage.PerformCalls).To(Equal([]*fakeui.PerformCall{
				{
					Name: "validating",
					Stage: &fakeui.FakeStage{
						PerformCalls: []*fakeui.PerformCall{
							{Name: "Validating deployment manifest"},
						},
					},
				},
				{
					Name:  "starting deployment",
					Stage: &fakeui.FakeStage{},
				},
			}))
		}

		Context("when the deployment state file does not exist", func() {
			BeforeEach(func() {
				err := fs.RemoveAll(deploymentStatePath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not starts anything", func() {
				err := newDeploymentStateManager().StartDeployment(fakeStage)
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

			It("starts the deployment", func() {
				expectStart()

				err := newDeploymentStateManager().StartDeployment(fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUI.Errors).To(BeEmpty())
			})

			It("logs validating & starting stages", func() {
				expectStart()

				err := newDeploymentStateManager().StartDeployment(fakeStage)
				Expect(err).ToNot(HaveOccurred())

				expectValidationStartEvents()
			})
		})

		Context("when nothing has been deployed", func() {
			BeforeEach(func() {
				err := setupDeploymentStateService.Save(biconfig.DeploymentState{DirectorID: "fake-uuid-0"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("tries to stop deployment", func() {
				expectStart()

				err := newDeploymentStateManager().StartDeployment(fakeStage)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeUI.Errors).To(BeEmpty())
			})
		})
	})
})
