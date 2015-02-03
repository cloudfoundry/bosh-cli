package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"time"

	"code.google.com/p/gomock/gomock"
	mock_blobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore/mocks"
	mock_agentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/mocks"
	mock_instance_state "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state/mocks"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Manager", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		fakeCloud *fakebmcloud.FakeCloud

		mockStateBuilderFactory *mock_instance_state.MockBuilderFactory
		mockStateBuilder        *mock_instance_state.MockBuilder
		mockState               *mock_instance_state.MockState

		mockBlobstore *mock_blobstore.MockBlobstore

		fakeVMManager        *fakebmvm.FakeManager
		fakeSSHTunnelFactory *fakebmsshtunnel.FakeFactory
		fakeSSHTunnel        *fakebmsshtunnel.FakeTunnel
		instanceFactory      Factory
		logger               boshlog.Logger
		fakeStage            *fakebmlog.FakeStage

		manager Manager
	)

	BeforeEach(func() {
		fakeCloud = fakebmcloud.NewFakeCloud()

		fakeVMManager = fakebmvm.NewFakeManager()

		fakeSSHTunnelFactory = fakebmsshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebmsshtunnel.NewFakeTunnel()
		fakeSSHTunnel.SetStartBehavior(nil, nil)
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel

		mockStateBuilderFactory = mock_instance_state.NewMockBuilderFactory(mockCtrl)
		mockStateBuilder = mock_instance_state.NewMockBuilder(mockCtrl)
		mockState = mock_instance_state.NewMockState(mockCtrl)

		instanceFactory = NewFactory(mockStateBuilderFactory)

		mockBlobstore = mock_blobstore.NewMockBlobstore(mockCtrl)

		logger = boshlog.NewLogger(boshlog.LevelNone)

		fakeStage = fakebmlog.NewFakeStage()

		manager = NewManager(
			fakeCloud,
			fakeVMManager,
			mockBlobstore,
			fakeSSHTunnelFactory,
			instanceFactory,
			logger,
		)
	})

	Describe("Create", func() {
		var (
			mockAgentClient    *mock_agentclient.MockAgentClient
			fakeVM             *fakebmvm.FakeVM
			diskPool           bmdeplmanifest.DiskPool
			deploymentManifest bmdeplmanifest.Manifest
			fakeCloudStemcell  *fakebmstemcell.FakeCloudStemcell
			registry           bminstallmanifest.Registry
			sshTunnelConfig    bminstallmanifest.SSHTunnel

			expectedInstance Instance
			expectedDisk     *fakebmdisk.FakeDisk
		)

		var allowApplySpecToBeCreated = func() {
			jobName := "cpi"
			jobIndex := 0

			applySpec := bmas.ApplySpec{
				Deployment: "test-release",
				Index:      jobIndex,
				Packages:   map[string]bmas.Blob{},
				Networks: map[string]bmproperty.Map{
					"network-1": bmproperty.Map{
						"cloud_properties": bmproperty.Map{},
						"type":             "dynamic",
						"ip":               "",
					},
				},
				Job: bmas.Job{
					Name:      jobName,
					Templates: []bmas.Blob{},
				},
				RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{},
				ConfigurationHash:        "",
			}

			mockStateBuilderFactory.EXPECT().NewBuilder(mockBlobstore, mockAgentClient).Return(mockStateBuilder).AnyTimes()
			mockStateBuilder.EXPECT().Build(jobName, jobIndex, deploymentManifest, fakeStage).Return(mockState, nil).AnyTimes()
			mockState.EXPECT().ToApplySpec().Return(applySpec).AnyTimes()
		}

		BeforeEach(func() {
			diskPool = bmdeplmanifest.DiskPool{
				Name:     "fake-persistent-disk-pool-name",
				DiskSize: 1024,
				CloudProperties: bmproperty.Map{
					"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
				},
			}

			deploymentManifest = bmdeplmanifest.Manifest{
				Update: bmdeplmanifest.Update{
					UpdateWatchTime: bmdeplmanifest.WatchTime{
						Start: 0,
						End:   5478,
					},
				},
				DiskPools: []bmdeplmanifest.DiskPool{
					diskPool,
				},
				Jobs: []bmdeplmanifest.Job{
					{
						Name:               "fake-job-name",
						PersistentDiskPool: "fake-persistent-disk-pool-name",
						Instances:          1,
					},
				},
			}

			fakeCloudStemcell = fakebmstemcell.NewFakeCloudStemcell("fake-stemcell-cid", "fake-stemcell-name", "fake-stemcell-version")
			registry = bminstallmanifest.Registry{}
			sshTunnelConfig = bminstallmanifest.SSHTunnel{}

			fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")
			fakeVMManager.CreateVM = fakeVM

			mockAgentClient = mock_agentclient.NewMockAgentClient(mockCtrl)
			fakeVM.AgentClientReturn = mockAgentClient

			expectedInstance = NewInstance(
				"fake-job-name",
				0,
				fakeVM,
				fakeVMManager,
				fakeSSHTunnelFactory,
				mockStateBuilder,
				logger,
			)

			expectedDisk = fakebmdisk.NewFakeDisk("fake-disk-cid")
			fakeVM.UpdateDisksDisks = []bmdisk.Disk{expectedDisk}
		})

		JustBeforeEach(func() {
			allowApplySpecToBeCreated()
		})

		It("returns an Instance that wraps a newly created VM", func() {
			instance, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				registry,
				sshTunnelConfig,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).To(Equal(expectedInstance))

			Expect(fakeVMManager.CreateInput).To(Equal(fakebmvm.CreateInput{
				Stemcell: fakeCloudStemcell,
				Manifest: deploymentManifest,
			}))
		})

		It("updates the current stemcell", func() {
			_, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				registry,
				sshTunnelConfig,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeCloudStemcell.PromoteAsCurrentCalledTimes).To(Equal(1))
		})

		It("logs start and stop events to the eventLogger", func() {
			_, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				registry,
				sshTunnelConfig,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Creating VM for instance 'fake-job-name/0' from stemcell 'fake-stemcell-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when registry settings are empty", func() {
			BeforeEach(func() {
				registry = bminstallmanifest.Registry{}
			})

			It("does not start the registry", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					registry,
					sshTunnelConfig,
					fakeStage,
				)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		It("waits for the vm", func() {
			_, _, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				registry,
				sshTunnelConfig,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.WaitUntilReadyInputs).To(Equal([]fakebmvm.WaitUntilReadyInput{
				{
					Timeout: 10 * time.Minute,
					Delay:   500 * time.Millisecond,
				},
			}))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		It("returns the 'updated' disks", func() {
			_, disks, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				registry,
				sshTunnelConfig,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(disks).To(Equal([]bmdisk.Disk{expectedDisk}))

			Expect(fakeVM.UpdateDisksInputs).To(Equal([]fakebmvm.UpdateDisksInput{
				{
					DiskPool: diskPool,
					Stage:    fakeStage,
				},
			}))
		})

		Context("when registry or sshTunnelConfig are not empty", func() {
			BeforeEach(func() {
				registry = bminstallmanifest.Registry{
					Username: "fake-registry-username",
					Password: "fake-registry-password",
					Host:     "fake-registry-host",
					Port:     124,
				}
				sshTunnelConfig = bminstallmanifest.SSHTunnel{
					User:       "fake-ssh-user",
					Host:       "fake-ssh-host",
					Port:       123,
					Password:   "fake-ssh-password",
					PrivateKey: "fake-ssh-private-key-path",
				}
			})

			It("starts & stops the ssh tunnel", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					registry,
					sshTunnelConfig,
					fakeStage,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bmsshtunnel.Options{
					User:              "fake-ssh-user",
					Host:              "fake-ssh-host",
					Port:              123,
					Password:          "fake-ssh-password",
					PrivateKey:        "fake-ssh-private-key-path",
					LocalForwardPort:  124,
					RemoteForwardPort: 124,
				}))
				Expect(fakeSSHTunnel.Started).To(BeTrue())
				Expect(fakeSSHTunnel.Stopped).To(BeTrue())
			})

			Context("when starting the ssh tunnel fails", func() {
				BeforeEach(func() {
					fakeSSHTunnel.SetStartBehavior(errors.New("fake-ssh-tunnel-start-error"), nil)
				})

				It("returns an error", func() {
					_, _, err := manager.Create(
						"fake-job-name",
						0,
						deploymentManifest,
						fakeCloudStemcell,
						registry,
						sshTunnelConfig,
						fakeStage,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
				})
			})
		})

		Context("when ssh tunnel conifg is empty", func() {
			BeforeEach(func() {
				sshTunnelConfig = bminstallmanifest.SSHTunnel{}
			})

			It("does not start the ssh tunnel", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					registry,
					sshTunnelConfig,
					fakeStage,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeSSHTunnel.Started).To(BeFalse())
				Expect(fakeSSHTunnel.Stopped).To(BeFalse())
			})
		})

		Context("when creating VM fails", func() {
			BeforeEach(func() {
				fakeVMManager.CreateErr = errors.New("fake-create-vm-error")
			})

			It("returns an error", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					registry,
					sshTunnelConfig,
					fakeStage,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-vm-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				_, _, err := manager.Create(
					"fake-job-name",
					0,
					deploymentManifest,
					fakeCloudStemcell,
					registry,
					sshTunnelConfig,
					fakeStage,
				)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Creating VM for instance 'fake-job-name/0' from stemcell 'fake-stemcell-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Creating VM: fake-create-vm-error",
				}))
			})
		})
	})
})
