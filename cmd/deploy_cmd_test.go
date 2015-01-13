package cmd_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_cloud "github.com/cloudfoundry/bosh-micro-cli/cloud/mocks"
	mock_httpagent "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/http/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"
	mock_deployer "github.com/cloudfoundry/bosh-micro-cli/deployment/mocks"
	mock_vm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/mocks"
	mock_install "github.com/cloudfoundry/bosh-micro-cli/installation/mocks"
	mock_registry "github.com/cloudfoundry/bosh-micro-cli/registry/mocks"
	mock_release "github.com/cloudfoundry/bosh-micro-cli/release/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstall "github.com/cloudfoundry/bosh-micro-cli/installation"
	bminstalljob "github.com/cloudfoundry/bosh-micro-cli/installation/job"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"
	bmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/fakes"
	fakebmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest/validator/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakebmrelsetmanifest "github.com/cloudfoundry/bosh-micro-cli/release/set/manifest/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		command    bmcmd.Cmd
		userConfig bmconfig.UserConfig
		fakeFs     *fakesys.FakeFileSystem
		fakeUI     *fakeui.FakeUI

		mockDeploymentFactory     *mock_deployer.MockFactory
		mockInstaller             *mock_install.MockInstaller
		mockInstallerFactory      *mock_install.MockInstallerFactory
		mockReleaseExtractor      *mock_release.MockExtractor
		releaseManager            bmrel.Manager
		releaseSetResolver        bmrelset.Resolver
		mockRegistryServerManager *mock_registry.MockServerManager
		mockRegistryServer        *mock_registry.MockServer
		mockAgentClient           *mock_agentclient.MockAgentClient
		mockAgentClientFactory    *mock_httpagent.MockAgentClientFactory
		mockCloudFactory          *mock_cloud.MockFactory

		fakeCPIRelease        *fakebmrel.FakeRelease
		logger                boshlog.Logger
		mockVMManagerFactory  *mock_vm.MockManagerFactory
		fakeVMManager         *fakebmvm.FakeManager
		fakeStemcellExtractor *fakebmstemcell.FakeExtractor

		fakeDeployer         *fakebmdepl.FakeDeployer
		fakeDeploymentRecord *fakebmdepl.FakeDeploymentRecord

		fakeReleaseSetParser    *fakebmrelsetmanifest.FakeParser
		fakeInstallationParser  *fakebminstallmanifest.FakeParser
		fakeDeploymentParser    *fakebmdeplmanifest.FakeParser
		deploymentConfigService bmconfig.DeploymentConfigService
		fakeReleaseSetValidator *fakebmrelsetmanifest.FakeValidator
		fakeDeploymentValidator *fakebmdeplval.FakeValidator

		fakeCompressor    *fakecmd.FakeCompressor
		fakeJobRenderer   *fakebmtemp.FakeJobRenderer
		fakeUUIDGenerator *fakeuuid.FakeGenerator

		fakeEventLogger *fakebmlog.FakeEventLogger
		fakeStage       *fakebmlog.FakeStage

		deploymentManifestPath    string
		deploymentConfigPath      string
		cpiReleaseTarballPath     string
		stemcellTarballPath       string
		expectedExtractedStemcell bmstemcell.ExtractedStemcell
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		deploymentManifestPath = "/path/to/manifest.yml"
		deploymentConfigPath = "/path/to/deployment.json"
		userConfig = bmconfig.UserConfig{
			DeploymentManifestPath: deploymentManifestPath,
		}
		fakeFs.WriteFileString(deploymentManifestPath, "")

		mockDeploymentFactory = mock_deployer.NewMockFactory(mockCtrl)
		mockInstaller = mock_install.NewMockInstaller(mockCtrl)
		mockInstallerFactory = mock_install.NewMockInstallerFactory(mockCtrl)

		mockReleaseExtractor = mock_release.NewMockExtractor(mockCtrl)
		releaseManager = bmrel.NewManager(logger)
		releaseSetResolver = bmrelset.NewResolver(releaseManager, logger)

		mockRegistryServerManager = mock_registry.NewMockServerManager(mockCtrl)
		mockRegistryServer = mock_registry.NewMockServer(mockCtrl)

		mockAgentClientFactory = mock_httpagent.NewMockAgentClientFactory(mockCtrl)
		mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)
		mockAgentClientFactory.EXPECT().NewAgentClient(gomock.Any(), gomock.Any()).Return(mockAgentClient).AnyTimes()

		mockCloudFactory = mock_cloud.NewMockFactory(mockCtrl)

		mockVMManagerFactory = mock_vm.NewMockManagerFactory(mockCtrl)
		fakeVMManager = fakebmvm.NewFakeManager()
		mockVMManagerFactory.EXPECT().NewManager(gomock.Any(), mockAgentClient, gomock.Any()).Return(fakeVMManager).AnyTimes()

		fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()

		fakeDeployer = fakebmdepl.NewFakeDeployer()

		fakeReleaseSetParser = fakebmrelsetmanifest.NewFakeParser()
		fakeInstallationParser = fakebminstallmanifest.NewFakeParser()
		fakeDeploymentParser = fakebmdeplmanifest.NewFakeParser()

		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService(deploymentConfigPath, fakeFs, fakeUUIDGenerator, logger)

		fakeReleaseSetValidator = fakebmrelsetmanifest.NewFakeValidator()
		fakeDeploymentValidator = fakebmdeplval.NewFakeValidator()

		fakeEventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior(fakeStage)

		fakeCompressor = fakecmd.NewFakeCompressor()
		fakeJobRenderer = fakebmtemp.NewFakeJobRenderer()

		fakeDeploymentRecord = fakebmdepl.NewFakeDeploymentRecord()

		cpiReleaseTarballPath = "/release/tarball/path"

		stemcellTarballPath = "/stemcell/tarball/path"
		expectedExtractedStemcell = bmstemcell.NewExtractedStemcell(
			bmstemcell.Manifest{
				ImagePath:          "/stemcell/image/path",
				Name:               "fake-stemcell-name",
				Version:            "fake-stemcell-version",
				SHA1:               "fake-stemcell-sha1",
				RawCloudProperties: map[interface{}]interface{}{},
			},
			bmstemcell.ApplySpec{},
			"fake-extracted-path",
			fakeFs,
		)
	})

	JustBeforeEach(func() {
		command = bmcmd.NewDeployCmd(
			fakeUI,
			userConfig,
			fakeFs,
			fakeReleaseSetParser,
			fakeInstallationParser,
			fakeDeploymentParser,
			deploymentConfigService,
			fakeReleaseSetValidator,
			fakeDeploymentValidator,
			mockInstallerFactory,
			mockReleaseExtractor,
			releaseManager,
			releaseSetResolver,
			mockCloudFactory,
			mockAgentClientFactory,
			mockVMManagerFactory,
			fakeStemcellExtractor,
			fakeDeploymentRecord,
			mockDeploymentFactory,
			fakeEventLogger,
			logger,
		)
	})

	Describe("Run", func() {
		var (
			releaseSetManifest     bmrelsetmanifest.Manifest
			boshDeploymentManifest bmdeplmanifest.Manifest
			installationManifest   bminstallmanifest.Manifest
			cloud                  *fakebmcloud.FakeCloud

			directorID   = "fake-uuid-0"

			expectCPIReleaseExtract *gomock.Call
			expectInstall           *gomock.Call
			expectNewCloud          *gomock.Call
		)

		BeforeEach(func() {
			// create input files
			fakeFs.WriteFileString(cpiReleaseTarballPath, "")
			fakeFs.WriteFileString(stemcellTarballPath, "")

			// deployment is set
			userConfig.DeploymentManifestPath = deploymentManifestPath

			// deployment exists
			fakeFs.WriteFileString(userConfig.DeploymentManifestPath, "")

			// release set is valid
			fakeReleaseSetValidator.SetValidateBehavior([]fakebmrelsetmanifest.ValidateOutput{
				{Err: nil},
			})

			// deployment is valid
			fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
				{Err: nil},
			})

			// stemcell exists
			fakeFs.WriteFile(stemcellTarballPath, []byte{})

			releaseSetManifest = bmrelsetmanifest.Manifest{
				Releases: []bmrelmanifest.ReleaseRef{
					{
						Name:    "fake-cpi-release-name",
						Version: "1.0",
					},
				},
			}

			// parsed CPI deployment manifest
			installationManifest = bminstallmanifest.Manifest{
				Release:  "fake-cpi-release-name",
				Registry: bminstallmanifest.Registry{},
				SSHTunnel: bminstallmanifest.SSHTunnel{
					Host: "fake-host",
				},
				Mbus: "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint",
			}

			// parsed BOSH deployment manifest
			boshDeploymentManifest = bmdeplmanifest.Manifest{
				Name: "fake-deployment-name",
				Jobs: []bmdeplmanifest.Job{
					{
						Name: "fake-job-name",
					},
				},
			}
			fakeDeploymentParser.ParseManifest = boshDeploymentManifest

			// parsed/extracted CPI release
			fakeCPIRelease = fakebmrel.NewFakeRelease()
			fakeCPIRelease.ReleaseName = "fake-cpi-release-name"
			fakeCPIRelease.ReleaseVersion = "1.0"
			fakeCPIRelease.ReleaseJobs = []bmrel.Job{
				{
					Name: "cpi",
					Templates: map[string]string{
						"templates/cpi.erb": "bin/cpi",
					},
				},
			}

			cloud = fakebmcloud.NewFakeCloud()
		})

		// allow return values of mocked methods to be modified by BeforeEach in child contexts
		JustBeforeEach(func() {
			fakeStemcellExtractor.SetExtractBehavior(stemcellTarballPath, expectedExtractedStemcell, nil)

			fakeReleaseSetParser.ParseManifest = releaseSetManifest
			fakeDeploymentParser.ParseManifest = boshDeploymentManifest
			fakeInstallationParser.ParseManifest = installationManifest

			fakeDeploymentRecord.SetIsDeployedBehavior(
				deploymentManifestPath,
				fakeCPIRelease,
				expectedExtractedStemcell,
				false,
				nil,
			)

			fakeDeploymentRecord.SetUpdateBehavior(
				deploymentManifestPath,
				fakeCPIRelease,
				nil,
			)

			installationPath := filepath.Join("fake-install-dir", "fake-installation-id")
			target := bminstall.NewTarget(installationPath)

			installedJob := bminstalljob.InstalledJob{
				Name: "cpi",
				Path: filepath.Join(target.JobsPath(), "cpi"),
			}

			mockInstallerFactory.EXPECT().NewInstaller().Return(mockInstaller, nil).AnyTimes()

			installation := bminstall.NewInstallation(target, installedJob, installationManifest, mockRegistryServerManager)

			expectInstall = mockInstaller.EXPECT().Install(installationManifest).Return(installation, nil).AnyTimes()

			deployment := bmdepl.NewDeployment(boshDeploymentManifest, fakeDeployer)
			mockDeploymentFactory.EXPECT().NewDeployment(boshDeploymentManifest).Return(deployment).AnyTimes()

			expectCPIReleaseExtract = mockReleaseExtractor.EXPECT().Extract(cpiReleaseTarballPath).Return(fakeCPIRelease, nil).AnyTimes()

			expectNewCloud = mockCloudFactory.EXPECT().NewCloud(installation, directorID).Return(cloud, nil).AnyTimes()

			fakeDeployer.SetDeployBehavior(nil)
		})

		It("prints the deployment manifest and state file", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeUI.Said).To(Equal([]string{
				"Deployment manifest: '/path/to/manifest.yml'",
				"Deployment state: '/path/to/deployment.json'",
			}))
		})

		It("adds a new event logger stage", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
				{Name: "validating"},
			}))

			Expect(fakeStage.Started).To(BeTrue())
			Expect(fakeStage.Finished).To(BeTrue())
		})

		It("parses the installation manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeInstallationParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("parses the deployment manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
		})

		It("validates release set manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleaseSetValidator.ValidateInputs).To(Equal([]fakebmrelsetmanifest.ValidateInput{
				{Manifest: releaseSetManifest},
			}))
		})

		It("validates bosh deployment manifest", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentValidator.ValidateInputs).To(Equal([]fakebmdeplval.ValidateInput{
				{Manifest: boshDeploymentManifest},
			}))
		})

		It("logs validation stages", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(Equal([]*fakebmlog.FakeStep{
				&fakebmlog.FakeStep{
					Name: "Validating stemcell",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmlog.FakeStep{
					Name: "Validating releases",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
				&fakebmlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				},
			}))
		})

		It("extracts CPI release tarball", func() {
			expectCPIReleaseExtract.Times(1)

			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		It("installs the CPI locally", func() {
			expectInstall.Times(1)
			expectNewCloud.Times(1)

			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the registry is configured", func() {
			BeforeEach(func() {
				installationManifest.Registry = bminstallmanifest.Registry{
					Username: "fake-username",
					Password: "fake-password",
					Host:     "fake-host",
					Port:     123,
				}
			})

			It("starts & stops the registry", func() {
				mockRegistryServerManager.EXPECT().Start("fake-username", "fake-password", "fake-host", 123).Return(mockRegistryServer, nil)
				mockRegistryServer.EXPECT().Stop()

				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("deletes the extracted CPI release", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCPIRelease.DeleteCalled).To(BeTrue())
		})

		It("extracts the stemcell", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeStemcellExtractor.ExtractInputs).To(Equal([]fakebmstemcell.ExtractInput{
				{TarballPath: stemcellTarballPath},
			}))
		})

		It("creates a VM", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeployer.DeployInputs).To(Equal([]fakebmdepl.DeployInput{
				{
					Cpi:             cloud,
					Manifest:        boshDeploymentManifest,
					Stemcell:        expectedExtractedStemcell,
					Registry:        installationManifest.Registry,
					SSHTunnelConfig: installationManifest.SSHTunnel,
					VMManager:       fakeVMManager,
				},
			}))
		})

		It("updates the deployment record", func() {
			err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentRecord.UpdateInputs).To(Equal([]fakebmdepl.UpdateInput{
				{
					ManifestPath: deploymentManifestPath,
					Release:      fakeCPIRelease,
				},
			}))
		})

		Context("when deployment has not changed", func() {
			JustBeforeEach(func() {
				fakeDeploymentRecord.SetIsDeployedBehavior(
					deploymentManifestPath,
					fakeCPIRelease,
					expectedExtractedStemcell,
					true,
					nil,
				)
			})

			It("skips deploy", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUI.Said).To(ContainElement("No deployment, stemcell or cpi release changes. Skipping deploy."))
				Expect(fakeDeployer.DeployInputs).To(BeEmpty())
			})
		})

		Context("when parsing the cpi deployment manifest fails", func() {
			BeforeEach(func() {
				fakeDeploymentParser.ParseErr = errors.New("fake-parse-error")
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing deployment manifest"))
				Expect(err.Error()).To(ContainSubstring("fake-parse-error"))
				Expect(fakeDeploymentParser.ParsePath).To(Equal(deploymentManifestPath))
			})
		})

		Context("when the cpi release does not contain a 'cpi' job", func() {
			BeforeEach(func() {
				fakeCPIRelease.ReleaseJobs = []bmrel.Job{
					{
						Name: "not-cpi",
					},
				}
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid CPI release 'fake-cpi-release-name': Job 'cpi' is missing from release"))
			})
		})

		Context("when multiple releases are given", func() {
			var (
				otherReleaseTarballPath string
				fakeOtherRelease        *fakebmrel.FakeRelease

				expectOtherReleaseExtract *gomock.Call
			)
			BeforeEach(func() {
				otherReleaseTarballPath = "/path/to/other-release.tgz"

				fakeFs.WriteFileString(otherReleaseTarballPath, "")

				fakeOtherRelease = fakebmrel.NewFakeRelease()
				fakeOtherRelease.ReleaseJobs = []bmrel.Job{
					{
						Name: "not-cpi",
					},
				}

				expectOtherReleaseExtract = mockReleaseExtractor.EXPECT().Extract(otherReleaseTarballPath).Return(fakeOtherRelease, nil).AnyTimes()
			})

			It("extracts all the release tarballs", func() {
				expectCPIReleaseExtract.Times(1)
				expectOtherReleaseExtract.Times(1)

				err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			It("installs the CPI release locally", func() {
				expectInstall.Times(1)
				expectNewCloud.Times(1)

				err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when cloud_provider.release is not specified", func() {
				BeforeEach(func() {
					installationManifest.Release = ""
				})

				It("returns error", func() {
					err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("cloud_provider.release must be provided"))
				})
			})

			Context("when cloud_provider.release refers to an undeclared release", func() {
				BeforeEach(func() {
					releaseSetManifest.Releases = []bmrelmanifest.ReleaseRef{}
				})

				It("uses the latest version of that release that is available", func() {
					expectInstall.Times(1)
					expectNewCloud.Times(1)

					err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when cloud_provider.release refers to an release declared with version 'latest'", func() {
				BeforeEach(func() {
					releaseSetManifest.Releases = []bmrelmanifest.ReleaseRef{
						{
							Name:    "fake-cpi-release-name",
							Version: "latest",
						},
					}
				})

				It("uses the latest version of that release that is available", func() {
					expectInstall.Times(1)
					expectNewCloud.Times(1)

					err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when cloud_provider.release is not a provided release", func() {
				BeforeEach(func() {
					installationManifest.Release = "missing-release"
				})

				It("returns error", func() {
					err := command.Run([]string{stemcellTarballPath, otherReleaseTarballPath, cpiReleaseTarballPath})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("cloud_provider.release 'missing-release' must refer to a provided release"))
				})
			})
		})

		Context("When the CPI release tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(cpiReleaseTarballPath)
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the release '/release/tarball/path' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating releases",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the release '/release/tarball/path' exists",
				}))
			})
		})

		Context("When the stemcell tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(stemcellTarballPath)
			})

			It("returns error", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the stemcell '/stemcell/tarball/path' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating stemcell",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the stemcell '/stemcell/tarball/path' exists",
				}))
			})
		})

		Context("when the deployment config file does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(deploymentConfigPath)
			})

			It("creates a deployment config", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := deploymentConfigService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig).To(Equal(bmconfig.DeploymentFile{
					DirectorID: directorID,
				}))
			})
		})

		Context("when there is no deployment set", func() {
			BeforeEach(func() {
				userConfig.DeploymentManifestPath = ""

				// re-create command to update userConfig.DeploymentFile
				command = bmcmd.NewDeployCmd(
					fakeUI,
					userConfig,
					fakeFs,
					fakeReleaseSetParser,
					fakeInstallationParser,
					fakeDeploymentParser,
					deploymentConfigService,
					fakeReleaseSetValidator,
					fakeDeploymentValidator,
					mockInstallerFactory,
					mockReleaseExtractor,
					releaseManager,
					releaseSetResolver,
					mockCloudFactory,
					mockAgentClientFactory,
					mockVMManagerFactory,
					fakeStemcellExtractor,
					fakeDeploymentRecord,
					mockDeploymentFactory,
					fakeEventLogger,
					logger,
				)
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Running deploy cmd: Deployment manifest not set"))
				Expect(fakeUI.Errors).To(ContainElement("Deployment manifest not set"))
			})
		})

		Context("when the deployment manifest does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll(deploymentManifestPath)
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Running deploy cmd: Deployment manifest does not exist at '/path/to/manifest.yml'"))
				Expect(fakeUI.Errors).To(ContainElement("Deployment manifest does not exist"))
			})
		})

		Context("when the deployment manifest is invalid", func() {
			BeforeEach(func() {
				fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
					{Err: errors.New("fake-deployment-validation-error")},
				})
			})

			It("returns err", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-deployment-validation-error"))
			})

			It("logs the failed event log", func() {
				err := command.Run([]string{stemcellTarballPath, cpiReleaseTarballPath})
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Validating deployment manifest: fake-deployment-validation-error",
				}))
			})
		})

		It("returns err when no arguments are given", func() {
			err := command.Run([]string{})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires at least 2 arguments"))
		})

		It("returns err when 1 argument is given", func() {
			err := command.Run([]string{"something"})
			Expect(err).To(HaveOccurred())
			Expect(fakeUI.Errors).To(ContainElement("Invalid usage - deploy command requires at least 2 arguments"))
		})
	})
})
