package instance_test

import (
	. "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
)

var _ = Describe("Manager", func() {
	var (
		fakeCloud *fakebmcloud.FakeCloud

		fakeVMManager        *fakebmvm.FakeManager
		fakeSSHTunnelFactory *fakebmsshtunnel.FakeFactory
		fakeSSHTunnel        *fakebmsshtunnel.FakeTunnel
		fakeDiskDeployer     *fakebmvm.FakeDiskDeployer
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

		fakeDiskDeployer = fakebmvm.NewFakeDiskDeployer()

		logger = boshlog.NewLogger(boshlog.LevelNone)

		fakeStage = fakebmlog.NewFakeStage()

		manager = NewManager(
			fakeCloud,
			fakeVMManager,
			fakeSSHTunnelFactory,
			logger,
		)
	})

	Describe("Create", func() {
		var (
			fakeVM             *fakebmvm.FakeVM
			diskPool           bmdeplmanifest.DiskPool
			deploymentManifest bmdeplmanifest.Manifest
			extractedStemcell  bmstemcell.ExtractedStemcell
			fakeCloudStemcell  *fakebmstemcell.FakeCloudStemcell
			registry           bminstallmanifest.Registry
			sshTunnelConfig    bminstallmanifest.SSHTunnel

			expectedInstance Instance
		)

		BeforeEach(func() {
			diskPool = bmdeplmanifest.DiskPool{
				Name:     "fake-persistent-disk-pool-name",
				DiskSize: 1024,
				RawCloudProperties: map[interface{}]interface{}{
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

			applySpec := bmstemcell.ApplySpec{
				Job: bmstemcell.Job{
					Name: "fake-job-name",
				},
			}
			fakeFs := fakesys.NewFakeFileSystem()
			extractedStemcell = bmstemcell.NewExtractedStemcell(
				bmstemcell.Manifest{},
				applySpec,
				"fake-extracted-path",
				fakeFs,
			)

			fakeCloudStemcell = fakebmstemcell.NewFakeCloudStemcell("fake-stemcell-cid", "fake-stemcell-name", "fake-stemcell-version")
			registry = bminstallmanifest.Registry{}
			sshTunnelConfig = bminstallmanifest.SSHTunnel{}

			fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")
			fakeVMManager.CreateVM = fakeVM

			expectedInstance = NewInstance(
				"fake-job-name",
				0,
				fakeVM,
				fakeVMManager,
				fakeSSHTunnelFactory,
				logger,
			)
		})

		It("creates a VM", func() {
			instance, err := manager.Create(
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
			instance, err := manager.Create(
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

			Expect(fakeCloudStemcell.PromoteAsCurrentCalledTimes).To(Equal(1))
		})

		It("logs start and stop events to the eventLogger", func() {
			instance, err := manager.Create(
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
				_, err := manager.Create(
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
			_, err := manager.Create(
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

		It("updates the disks", func() {
			_, err := manager.Create(
				"fake-job-name",
				0,
				deploymentManifest,
				fakeCloudStemcell,
				registry,
				sshTunnelConfig,
				fakeStage,
			)
			Expect(err).NotTo(HaveOccurred())

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
				_, err := manager.Create(
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
					_, err := manager.Create(
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
				_, err := manager.Create(
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
				_, err := manager.Create(
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
				_, err := manager.Create(
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
