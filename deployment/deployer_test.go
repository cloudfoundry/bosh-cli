package deployer_test

import (
	"errors"
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmretry "github.com/cloudfoundry/bosh-micro-cli/deployment/retrystrategy/fakes"
	fakebmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Deployer", func() {

	var (
		deployer                   Deployer
		fakeVMManagerFactory       *fakebmvm.FakeManagerFactory
		fakeVMManager              *fakebmvm.FakeManager
		fakeSSHTunnelFactory       *fakebmsshtunnel.FakeFactory
		fakeSSHTunnel              *fakebmsshtunnel.FakeTunnel
		fakeDiskDeployer           *fakebmvm.FakeDiskDeployer
		cloud                      *fakebmcloud.FakeCloud
		deploymentManifest         bmmanifest.Manifest
		diskPool                   bmmanifest.DiskPool
		registryConfig             bmmanifest.Registry
		eventLogger                *fakebmlog.FakeEventLogger
		fakeStage                  *fakebmlog.FakeStage
		sshTunnelConfig            bmmanifest.SSHTunnel
		fakeAgentPingRetryStrategy *fakebmretry.FakeRetryStrategy
		fakeVM                     *fakebmvm.FakeVM
		fakeStemcellManager        *fakebmstemcell.FakeManager
		fakeStemcellManagerFactory *fakebmstemcell.FakeManagerFactory
		extractedStemcell          bmstemcell.ExtractedStemcell

		applySpec     bmstemcell.ApplySpec
		cloudStemcell bmstemcell.CloudStemcell
	)

	BeforeEach(func() {
		diskPool = bmmanifest.DiskPool{
			Name:     "fake-persistent-disk-pool-name",
			DiskSize: 1024,
			RawCloudProperties: map[interface{}]interface{}{
				"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
			},
		}
		deploymentManifest = bmmanifest.Manifest{
			Update: bmmanifest.Update{
				UpdateWatchTime: bmmanifest.WatchTime{
					Start: 0,
					End:   5478,
				},
			},
			DiskPools: []bmmanifest.DiskPool{
				diskPool,
			},
			Jobs: []bmmanifest.Job{
				{
					Name:               "fake-job-name",
					PersistentDiskPool: "fake-persistent-disk-pool-name",
					Instances:          1,
				},
			},
		}
		registryConfig = bmmanifest.Registry{}
		sshTunnelConfig = bmmanifest.SSHTunnel{
			User:       "fake-ssh-username",
			PrivateKey: "fake-private-key-path",
			Password:   "fake-password",
			Host:       "fake-ssh-host",
			Port:       124,
		}

		cloud = fakebmcloud.NewFakeCloud()

		fakeVMManagerFactory = fakebmvm.NewFakeManagerFactory()
		fakeVMManager = fakebmvm.NewFakeManager()
		fakeVMManagerFactory.SetNewManagerBehavior(cloud, "fake-mbus-url", fakeVMManager)

		fakeSSHTunnelFactory = fakebmsshtunnel.NewFakeFactory()
		fakeSSHTunnel = fakebmsshtunnel.NewFakeTunnel()
		fakeSSHTunnelFactory.SSHTunnel = fakeSSHTunnel
		fakeSSHTunnel.SetStartBehavior(nil, nil)

		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")
		fakeVMManager.CreateVM = fakeVM

		fakeDiskDeployer = fakebmvm.NewFakeDiskDeployer()

		logger := boshlog.NewLogger(boshlog.LevelNone)
		eventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		eventLogger.SetNewStageBehavior(fakeStage)

		fakeAgentPingRetryStrategy = fakebmretry.NewFakeRetryStrategy()

		applySpec = bmstemcell.ApplySpec{
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

		fakeStemcellManager = fakebmstemcell.NewFakeManager()
		fakeStemcellManagerFactory = fakebmstemcell.NewFakeManagerFactory()
		fakeStemcellRepo := fakebmconfig.NewFakeStemcellRepo()
		stemcellRecord := bmconfig.StemcellRecord{
			ID:      "fake-stemcell-id",
			Name:    "fake-stemcell-name",
			Version: "fake-stemcell-version",
			CID:     "fake-stemcell-cid",
		}
		err := fakeStemcellRepo.SetFindBehavior("fake-stemcell-name", "fake-stemcell-version", stemcellRecord, true, nil)
		Expect(err).ToNot(HaveOccurred())

		cloudStemcell = bmstemcell.NewCloudStemcell(stemcellRecord, fakeStemcellRepo, cloud)
		fakeStemcellManager.SetUploadBehavior(extractedStemcell, cloudStemcell, nil)
		fakeStemcellManagerFactory.SetNewManagerBehavior(cloud, fakeStemcellManager)

		deployer = NewDeployer(
			fakeStemcellManagerFactory,
			fakeVMManagerFactory,
			fakeSSHTunnelFactory,
			eventLogger,
			logger,
		)
	})

	It("uploads the stemcell", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeStemcellManager.UploadInputs).To(Equal([]fakebmstemcell.UploadInput{
			{Stemcell: extractedStemcell},
		}))
	})

	It("adds a new event logger stage", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).ToNot(HaveOccurred())

		Expect(eventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
			{Name: "deploying"},
		}))

		Expect(fakeStage.Started).To(BeTrue())
		Expect(fakeStage.Finished).To(BeTrue())
	})

	Context("when a previous instance exists", func() {
		var fakeExistingVM *fakebmvm.FakeVM

		BeforeEach(func() {
			fakeExistingVM = fakebmvm.NewFakeVM("existing-vm-cid")
			fakeVMManager.SetFindCurrentBehavior(fakeExistingVM, true, nil)
		})

		It("deletes existing vm", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExistingVM.DeleteCalled).To(Equal(1))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for the agent on VM 'existing-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Deleting VM 'existing-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})
	})

	It("creates a vm", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVMManager.CreateInput).To(Equal(fakebmvm.CreateInput{
			Stemcell: cloudStemcell,
			Manifest: deploymentManifest,
		}))
	})

	It("deletes unused stemcells", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStemcellManager.DeleteUnusedCalledTimes).To(Equal(1))
	})

	Context("when registry & ssh tunnel configs are not empty", func() {
		BeforeEach(func() {
			registryConfig = bmmanifest.Registry{
				Username: "fake-username",
				Password: "fake-password",
				Host:     "fake-host",
				Port:     123,
			}
			sshTunnelConfig = bmmanifest.SSHTunnel{
				User:       "fake-ssh-username",
				PrivateKey: "fake-private-key-path",
				Password:   "fake-password",
				Host:       "fake-ssh-host",
				Port:       124,
			}
		})

		It("starts the SSH tunnel", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeSSHTunnel.Started).To(BeTrue())
			Expect(fakeSSHTunnelFactory.NewSSHTunnelOptions).To(Equal(bmsshtunnel.Options{
				User:              "fake-ssh-username",
				PrivateKey:        "fake-private-key-path",
				Password:          "fake-password",
				Host:              "fake-ssh-host",
				Port:              124,
				LocalForwardPort:  123,
				RemoteForwardPort: 123,
			}))
		})

		Context("when starting SSH tunnel fails", func() {
			BeforeEach(func() {
				fakeSSHTunnel.SetStartBehavior(errors.New("fake-ssh-tunnel-start-error"), nil)
			})

			It("returns an error", func() {
				err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-ssh-tunnel-start-error"))
			})
		})
	})

	It("waits for the vm", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeVM.WaitUntilReadyInputs).To(ContainElement(fakebmvm.WaitUntilReadyInput{
			Timeout: 10 * time.Minute,
			Delay:   500 * time.Millisecond,
		}))
	})

	It("logs start and stop events to the eventLogger", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
	})

	Context("when waiting for the agent fails", func() {
		BeforeEach(func() {
			fakeVM.WaitUntilReadyErr = errors.New("fake-wait-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for the agent on VM 'fake-vm-cid' to be ready",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-wait-error",
			}))
		})
	})

	It("updates the vm", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.ApplyInputs).To(ContainElement(fakebmvm.ApplyInput{
			StemcellApplySpec: applySpec,
			Manifest:          deploymentManifest,
		}))
	})

	It("starts the agent", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.StartCalled).To(Equal(1))
	})

	It("waits until agent reports state as running", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebmvm.WaitInput{
			MaxAttempts: 5,
			Delay:       1 * time.Second,
		}))
	})

	Context("when the deployment has an invalid disk pool specification", func() {
		BeforeEach(func() {
			deploymentManifest.Jobs[0].PersistentDiskPool = "fake-non-existent-persistent-disk-pool-name"
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
		})
	})

	It("logs start and stop events to the eventLogger", func() {
		err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Starting instance 'fake-job-name/0'",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Waiting for instance 'fake-job-name/0' to be running",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
	})

	Context("when uploading stemcell fails", func() {
		BeforeEach(func() {
			fakeStemcellManager.SetUploadBehavior(extractedStemcell, nil, errors.New("fake-upload-error"))
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-upload-error"))
		})
	})

	Context("when updating instance fails", func() {
		BeforeEach(func() {
			fakeVM.ApplyErr = errors.New("fake-apply-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-apply-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Starting instance 'fake-job-name/0'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "Applying the agent state: fake-apply-error",
			}))
		})
	})

	Context("when starting agent services fails", func() {
		BeforeEach(func() {
			fakeVM.StartErr = errors.New("fake-start-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-start-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Starting instance 'fake-job-name/0'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "Starting the agent: fake-start-error",
			}))
		})
	})

	Context("when waiting for running state fails", func() {
		BeforeEach(func() {
			fakeVM.WaitToBeRunningErr = errors.New("fake-wait-running-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deploymentManifest, extractedStemcell, registryConfig, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for instance 'fake-job-name/0' to be running",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "fake-wait-running-error",
			}))
		})
	})
})
