package cmd_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gomock/gomock"
	mock_cloud "github.com/cloudfoundry/bosh-micro-cli/cloud/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"

	fakebmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi/fakes"
)

var _ = Describe("Cmd/DeleteCmd", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Run", func() {
		var (
			fs                      boshsys.FileSystem
			logger                  boshlog.Logger
			fakeCPIInstaller        *fakebmcpi.FakeInstaller
			uuidGenerator           boshuuid.Generator
			deploymentConfigService bmconfig.DeploymentConfigService
			vmRepo                  bmconfig.VMRepo
			diskRepo                bmconfig.DiskRepo
			stemcellRepo            bmconfig.StemcellRepo
			userConfig              bmconfig.UserConfig

			ui *fakeui.FakeUI

			mockAgentClient        *mock_agentclient.MockAgentClient
			mockAgentClientFactory *mock_agentclient.MockFactory
			mockCloud              *mock_cloud.MockCloud
		)

		var writeDeploymentManifest = func() {
			fs.WriteFileString("/deployment-dir/fake-deployment-manifest.yml", `---
name: test-release

cloud_provider:
  mbus: http://fake-mbus-url
`)
		}

		var writeCPIReleaseTarball = func() {
			fs.WriteFileString("/fake-cpi-release.tgz", "fake-tgz-content")
		}

		var allowCPIToBeExtracted = func() {
			cpiRelease := bmrel.NewRelease(
				"fake-cpi-release-name",
				"fake-cpi-release-version",
				[]bmrel.Job{},
				[]*bmrel.Package{},
				"fake-extracted-dir",
				fs,
			)
			fakeCPIInstaller.SetExtractBehavior("/fake-cpi-release.tgz", cpiRelease, nil)
		}

		var allowCPIToBeInstalled = func() {
			cpiRelease := bmrel.NewRelease(
				"fake-cpi-release-name",
				"fake-cpi-release-version",
				[]bmrel.Job{},
				[]*bmrel.Package{},
				"fake-extracted-dir",
				fs,
			)
			cpiDeployment := bmdepl.CPIDeployment{
				Name: "test-release",
				Mbus: "http://fake-mbus-url",
			}
			fakeCPIInstaller.SetInstallBehavior(cpiDeployment, cpiRelease, mockCloud, nil)
		}

		var newDeleteCmd = func() Cmd {
			deploymentParser := bmdepl.NewParser(fs, logger)
			eventLogger := bmeventlog.NewEventLogger(ui)
			return NewDeleteCmd(
				ui, userConfig, fs, deploymentParser, fakeCPIInstaller,
				vmRepo, diskRepo, stemcellRepo, mockAgentClientFactory,
				eventLogger, logger,
			)
		}

		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			logger = boshlog.NewLogger(boshlog.LevelNone)
			deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService("/fake-bosh-deployments.json", fs, logger)
			uuidGenerator = boshuuid.Generator(nil)

			vmRepo = bmconfig.NewVMRepo(deploymentConfigService)
			diskRepo = bmconfig.NewDiskRepo(deploymentConfigService, uuidGenerator)
			stemcellRepo = bmconfig.NewStemcellRepo(deploymentConfigService, uuidGenerator)

			mockCloud = mock_cloud.NewMockCloud(mockCtrl)

			fakeCPIInstaller = fakebmcpi.NewFakeInstaller()

			ui = &fakeui.FakeUI{}

			mockAgentClientFactory = mock_agentclient.NewMockFactory(mockCtrl)
			mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)

			userConfig = bmconfig.UserConfig{
				DeploymentFile: "/deployment-dir/fake-deployment-manifest.yml",
			}

			writeDeploymentManifest()
			writeCPIReleaseTarball()
			allowCPIToBeExtracted()
			allowCPIToBeInstalled()
		})

		Context("when the deployment has not been set", func() {
			BeforeEach(func() {
				userConfig.DeploymentFile = ""
			})

			It("returns an error", func() {
				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No deployment set"))
			})
		})

		Context("when microbosh has been deployed", func() {
			BeforeEach(func() {
				// create deployment manifest yaml file
				deploymentConfigService.Save(bmconfig.DeploymentFile{
					UUID:              "",
					CurrentVMCID:      "fake-vm-cid",
					CurrentStemcellID: "fake-stemcell-guid",
					CurrentDiskID:     "fake-disk-guid",
					Disks: []bmconfig.DiskRecord{
						{
							ID:   "fake-disk-guid",
							CID:  "fake-disk-cid",
							Size: 100,
						},
					},
					Stemcells: []bmconfig.StemcellRecord{
						{
							ID:  "fake-stemcell-guid",
							CID: "fake-stemcell-cid",
						},
					},
				})
			})

			It("stops the agent, then deletes the vm, disk, and stemcell", func() {
				gomock.InOrder(
					mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
					mockAgentClient.EXPECT().Stop(),
					mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
					mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
					mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
				)

				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).ToNot(HaveOccurred())
			})

			It("logs validation stages", func() {
				gomock.InOrder(
					mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
					mockAgentClient.EXPECT().Stop(),
					mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
					mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
					mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
				)

				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Said).To(ConsistOf(
					"Started validating",
					"Started validating > Validating deployment manifest...", " done. (00:00:00)",
					"Started validating > Validating cpi release...", " done. (00:00:00)",
					"Done validating",
					"",
					// if cpiInstaller were not mocked, it would print the "installing CPI jobs" stage here.
					"Started deleting deployment",
					"Started deleting deployment > Stopping agent...", " done. (00:00:00)",
					"Started deleting deployment > Deleting VM...", " done. (00:00:00)",
					"Started deleting deployment > Deleting disk...", " done. (00:00:00)",
					"Started deleting deployment > Deleting stemcell...", " done. (00:00:00)",
					"Done deleting deployment",
					"",
				))
			})

			It("clears current vm, disk and stemcell", func() {
				gomock.InOrder(
					mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
					mockAgentClient.EXPECT().Stop(),
					mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
					mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
					mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
				)

				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).ToNot(HaveOccurred())

				_, found, err := vmRepo.FindCurrent()
				Expect(found).To(BeFalse(), "should be no current VM")

				_, found, err = diskRepo.FindCurrent()
				Expect(found).To(BeFalse(), "should be no current disk")

				diskRecords, err := diskRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(diskRecords).To(BeEmpty(), "expected no disk records")

				_, found, err = stemcellRepo.FindCurrent()
				Expect(found).To(BeFalse(), "should be no current stemcell")

				stemcellRecords, err := stemcellRepo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
			})

			Context("and orphan disks exist", func() {
				BeforeEach(func() {
					deploymentFile, err := deploymentConfigService.Load()
					Expect(err).ToNot(HaveOccurred())

					deploymentFile.Disks = append(deploymentFile.Disks, bmconfig.DiskRecord{
						ID:   "fake-disk-guid-2",
						CID:  "fake-disk-cid-2",
						Size: 1000,
					})

					err = deploymentConfigService.Save(deploymentFile)
					Expect(err).ToNot(HaveOccurred())
				})

				It("deletes the orphaned disks", func() {
					gomock.InOrder(
						mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
						mockAgentClient.EXPECT().Stop(),
						mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
						mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
						mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
					)

					mockCloud.EXPECT().DeleteDisk("fake-disk-cid-2")

					err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
					Expect(err).ToNot(HaveOccurred())

					diskRecords, err := diskRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(diskRecords).To(BeEmpty(), "expected no disk records")
				})

				It("logs validation stages", func() {
					gomock.InOrder(
						mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
						mockAgentClient.EXPECT().Stop(),
						mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
						mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
						mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
					)

					mockCloud.EXPECT().DeleteDisk("fake-disk-cid-2")

					err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(ConsistOf(
						"Started validating",
						"Started validating > Validating deployment manifest...", " done. (00:00:00)",
						"Started validating > Validating cpi release...", " done. (00:00:00)",
						"Done validating",
						"",
						// if cpiInstaller were not mocked, it would print the "installing CPI jobs" stage here.
						"Started deleting deployment",
						"Started deleting deployment > Stopping agent...", " done. (00:00:00)",
						"Started deleting deployment > Deleting VM...", " done. (00:00:00)",
						"Started deleting deployment > Deleting disk...", " done. (00:00:00)",
						"Started deleting deployment > Deleting stemcell...", " done. (00:00:00)",
						"Started deleting deployment > Deleting orphaned disks...", " done. (00:00:00)",
						"Done deleting deployment",
						"",
					))
				})
			})

			Context("and orphan stemcells exist", func() {
				BeforeEach(func() {
					deploymentFile, err := deploymentConfigService.Load()
					Expect(err).ToNot(HaveOccurred())

					deploymentFile.Stemcells = append(deploymentFile.Stemcells, bmconfig.StemcellRecord{
						ID:  "fake-stemcell-guid-2",
						CID: "fake-stemcell-cid-2",
					})

					err = deploymentConfigService.Save(deploymentFile)
					Expect(err).ToNot(HaveOccurred())
				})

				It("deletes the orphaned stemcells", func() {
					gomock.InOrder(
						mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
						mockAgentClient.EXPECT().Stop(),
						mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
						mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
						mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
					)

					mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid-2")

					err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
					Expect(err).ToNot(HaveOccurred())

					stemcellRecords, err := stemcellRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(stemcellRecords).To(BeEmpty(), "expected no stemcell records")
				})

				It("logs validation stages", func() {
					gomock.InOrder(
						mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
						mockAgentClient.EXPECT().Stop(),
						mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
						mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
						mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
					)

					mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid-2")

					err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Said).To(ConsistOf(
						"Started validating",
						"Started validating > Validating deployment manifest...", " done. (00:00:00)",
						"Started validating > Validating cpi release...", " done. (00:00:00)",
						"Done validating",
						"",
						// if cpiInstaller were not mocked, it would print the "installing CPI jobs" stage here.
						"Started deleting deployment",
						"Started deleting deployment > Stopping agent...", " done. (00:00:00)",
						"Started deleting deployment > Deleting VM...", " done. (00:00:00)",
						"Started deleting deployment > Deleting disk...", " done. (00:00:00)",
						"Started deleting deployment > Deleting stemcell...", " done. (00:00:00)",
						"Started deleting deployment > Deleting orphaned stemcells...", " done. (00:00:00)",
						"Done deleting deployment",
						"",
					))
				})
			})
		})

		Context("when microbosh has not been deployed", func() {
			BeforeEach(func() {
				deploymentConfigService.Save(bmconfig.DeploymentFile{
					UUID:              "",
					CurrentVMCID:      "",
					CurrentStemcellID: "",
					CurrentDiskID:     "",
				})
			})

			It("returns an error", func() {
				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No existing microbosh instance to delete"))
				Expect(ui.Errors).To(ContainElement("No existing microbosh instance to delete"))
			})
		})

		Context("when VM has been deployed", func() {
			BeforeEach(func() {
				deploymentConfigService.Save(bmconfig.DeploymentFile{
					UUID:              "",
					CurrentVMCID:      "fake-vm-cid",
					CurrentStemcellID: "",
					CurrentDiskID:     "",
				})
			})

			It("stops the agent and deletes the VM", func() {
				gomock.InOrder(
					mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
					mockAgentClient.EXPECT().Stop(),
					mockCloud.EXPECT().DeleteVM("fake-vm-cid"),
				)

				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when a current disk exists", func() {
			BeforeEach(func() {
				deploymentConfigService.Save(bmconfig.DeploymentFile{
					UUID:              "",
					CurrentVMCID:      "",
					CurrentStemcellID: "",
					CurrentDiskID:     "fake-disk-guid",
					Disks: []bmconfig.DiskRecord{
						{
							ID:   "fake-disk-guid",
							CID:  "fake-disk-cid",
							Size: 100,
						},
					},
				})
			})

			It("stops the agent and deletes the VM", func() {
				gomock.InOrder(
					mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
					mockAgentClient.EXPECT().Stop(),
					mockCloud.EXPECT().DeleteDisk("fake-disk-cid"),
				)

				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when a current stemcell exists", func() {
			BeforeEach(func() {
				deploymentConfigService.Save(bmconfig.DeploymentFile{
					UUID:              "",
					CurrentVMCID:      "",
					CurrentStemcellID: "fake-stemcell-guid",
					CurrentDiskID:     "",
					Stemcells: []bmconfig.StemcellRecord{
						{
							ID:  "fake-stemcell-guid",
							CID: "fake-stemcell-cid",
						},
					},
				})
			})

			It("stops the agent and deletes the VM", func() {
				gomock.InOrder(
					mockAgentClientFactory.EXPECT().Create("http://fake-mbus-url").Return(mockAgentClient),
					mockAgentClient.EXPECT().Stop(),
					mockCloud.EXPECT().DeleteStemcell("fake-stemcell-cid"),
				)

				err := newDeleteCmd().Run([]string{"/fake-cpi-release.tgz"})
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
