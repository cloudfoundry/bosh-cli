package deployer_test

import (
	"errors"
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmdeployer "github.com/cloudfoundry/bosh-micro-cli/deployer/fakes"
	fakeregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry/fakes"
	fakebmretry "github.com/cloudfoundry/bosh-micro-cli/deployer/retrystrategy/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("Deployer", func() {
	var (
		deployer                   Deployer
		fakeVMDeployer             *fakebmdeployer.FakeVMDeployer
		fakeDiskDeployer           *fakebmdeployer.FakeDiskDeployer
		cloud                      *fakebmcloud.FakeCloud
		deployment                 bmdepl.Deployment
		diskPool                   bmdepl.DiskPool
		registry                   bmdepl.Registry
		fakeRegistryServer         *fakeregistry.FakeServer
		eventLogger                *fakebmlog.FakeEventLogger
		fakeStage                  *fakebmlog.FakeStage
		sshTunnelConfig            bmdepl.SSHTunnel
		fakeAgentPingRetryStrategy *fakebmretry.FakeRetryStrategy
		fakeVM                     *fakebmvm.FakeVM
		fakeStemcellManager        *fakebmstemcell.FakeManager
		fakeStemcellManagerFactory *fakebmstemcell.FakeManagerFactory
		extractedStemcell          bmstemcell.ExtractedStemcell

		applySpec             bmstemcell.ApplySpec
		expectedCloudStemcell bmstemcell.CloudStemcell
	)

	BeforeEach(func() {
		diskPool = bmdepl.DiskPool{
			Name:     "fake-persistent-disk-pool-name",
			DiskSize: 1024,
			RawCloudProperties: map[interface{}]interface{}{
				"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
			},
		}
		deployment = bmdepl.Deployment{
			Update: bmdepl.Update{
				UpdateWatchTime: bmdepl.WatchTime{
					Start: 0,
					End:   5478,
				},
			},
			DiskPools: []bmdepl.DiskPool{
				diskPool,
			},
			Jobs: []bmdepl.Job{
				{
					Name:               "fake-job-name",
					PersistentDiskPool: "fake-persistent-disk-pool-name",
				},
			},
		}
		registry = bmdepl.Registry{
			Username: "fake-username",
			Password: "fake-password",
			Host:     "fake-host",
			Port:     123,
		}
		sshTunnelConfig = bmdepl.SSHTunnel{
			User:       "fake-ssh-username",
			PrivateKey: "fake-private-key-path",
			Password:   "fake-password",
			Host:       "fake-ssh-host",
			Port:       124,
		}

		cloud = fakebmcloud.NewFakeCloud()
		fakeRegistryServer = fakeregistry.NewFakeServer()

		fakeVMDeployer = fakebmdeployer.NewFakeVMDeployer()
		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")
		fakeVMDeployer.SetDeployBehavior(fakeVM, nil)

		fakeDiskDeployer = fakebmdeployer.NewFakeDiskDeployer()

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
		stemcellRepo := fakebmconfig.NewFakeStemcellRepo()
		stemcellRecord := bmconfig.StemcellRecord{}
		expectedCloudStemcell = bmstemcell.NewCloudStemcell(stemcellRecord, stemcellRepo, cloud)
		fakeStemcellManager.SetUploadBehavior(extractedStemcell, expectedCloudStemcell, nil)
		fakeStemcellManagerFactory.SetNewManagerBehavior(cloud, fakeStemcellManager)

		deployer = NewDeployer(
			fakeStemcellManagerFactory,
			fakeVMDeployer,
			fakeDiskDeployer,
			fakeRegistryServer,
			eventLogger,
			logger,
		)
	})

	It("uploads the stemcell", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).ToNot(HaveOccurred())
		Expect(fakeStemcellManager.UploadInputs).To(Equal(
			[]fakebmstemcell.UploadInput{
				{
					Stemcell: extractedStemcell,
				},
			},
		))
	})

	It("adds a new event logger stage", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).ToNot(HaveOccurred())

		Expect(eventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
			{
				Name: "deploying",
			},
		}))

		Expect(fakeStage.Started).To(BeTrue())
		Expect(fakeStage.Finished).To(BeTrue())
	})

	It("starts the registry", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeRegistryServer.StartInput).To(Equal(fakeregistry.StartInput{
			Username: "fake-username",
			Password: "fake-password",
			Host:     "fake-host",
			Port:     123,
		}))
		Expect(fakeRegistryServer.ReceivedActions).To(Equal([]string{"Start", "Stop"}))
	})

	Context("when registry settings are empty", func() {
		BeforeEach(func() {
			registry = bmdepl.Registry{}
		})

		It("starts the registry", func() {
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeRegistryServer.StartInput).To(Equal(fakeregistry.StartInput{}))
			Expect(fakeRegistryServer.ReceivedActions).To(BeEmpty())
		})
	})

	It("deploys vm", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVMDeployer.DeployInputs).To(ContainElement(fakebmdeployer.VMDeployInput{
			Cloud:            cloud,
			Deployment:       deployment,
			Stemcell:         expectedCloudStemcell,
			MbusURL:          "fake-mbus-url",
			EventLoggerStage: fakeStage,
		}))
	})

	It("deletes unused stemcells", func() {
		firstStemcell := fakebmstemcell.NewFakeCloudStemcell("fake-stemcell-cid-1", "fake-stemcell-name-1", "fake-stemcell-version-1")
		secondStemcell := fakebmstemcell.NewFakeCloudStemcell("fake-stemcell-cid-2", "fake-stemcell-name-2", "fake-stemcell-version-2")
		fakeStemcellManager.SetFindUnusedBehavior([]bmstemcell.CloudStemcell{
			firstStemcell,
			secondStemcell,
		}, nil)

		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(firstStemcell.DeleteCalledTimes).To(Equal(1))
		Expect(secondStemcell.DeleteCalledTimes).To(Equal(1))

		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Deleting unused stemcell 'fake-stemcell-cid-1'",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Deleting unused stemcell 'fake-stemcell-cid-2'",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
	})

	It("waits until vm is ready", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		expectedSSHTunnelOptions := bmsshtunnel.Options{
			Host:              "fake-ssh-host",
			Port:              124,
			User:              "fake-ssh-username",
			Password:          "fake-password",
			PrivateKey:        "fake-private-key-path",
			LocalForwardPort:  123,
			RemoteForwardPort: 123,
		}

		Expect(fakeVMDeployer.WaitUntilReadyInputs).To(ContainElement(fakebmdeployer.WaitUntilReadyInput{
			VM:               fakeVM,
			SSHTunnelOptions: expectedSSHTunnelOptions,
			EventLoggerStage: fakeStage,
		}))
	})

	It("updates the vm", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.ApplyInputs).To(ContainElement(fakebmvm.ApplyInput{
			StemcellApplySpec: applySpec,
			Deployment:        deployment,
		}))
	})

	It("deploys disk", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeDiskDeployer.DeployInputs).To(Equal([]fakebmdeployer.DiskDeployInput{
			{
				DiskPool:         diskPool,
				Cloud:            cloud,
				VM:               fakeVM,
				EventLoggerStage: fakeStage,
			},
		}))
	})

	It("starts the agent", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.StartCalled).To(Equal(1))
	})

	It("waits until agent reports state as running", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeVM.WaitToBeRunningInputs).To(ContainElement(fakebmvm.WaitInput{
			MaxAttempts: 5,
			Delay:       1 * time.Second,
		}))
	})

	Context("when the deployment has an invalid disk pool specification", func() {
		BeforeEach(func() {
			deployment.Jobs[0].PersistentDiskPool = "fake-non-existent-persistent-disk-pool-name"
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
		})
	})

	It("logs start and stop events to the eventLogger", func() {
		err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Starting 'fake-job-name'",
			States: []bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Finished,
			},
		}))
		Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
			Name: "Waiting for 'fake-job-name'",
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
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-upload-error"))
		})
	})

	Context("when starting registry fails", func() {
		BeforeEach(func() {
			fakeRegistryServer.SetStartBehavior(errors.New("fake-registry-start-error"), nil)
		})

		It("returns an error", func() {
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-registry-start-error"))
		})
	})

	Context("when updating instance fails", func() {
		BeforeEach(func() {
			fakeVM.ApplyErr = errors.New("fake-apply-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-apply-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Starting 'fake-job-name'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "Updating the vm: fake-apply-error",
			}))
		})
	})

	Context("when starting agent services fails", func() {
		BeforeEach(func() {
			fakeVM.StartErr = errors.New("fake-start-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-start-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Starting 'fake-job-name'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "Starting vm: fake-start-error",
			}))
		})
	})

	Context("when waiting for running state fails", func() {
		BeforeEach(func() {
			fakeVM.WaitToBeRunningErr = errors.New("fake-wait-running-error")
		})

		It("logs start and stop events to the eventLogger", func() {
			err := deployer.Deploy(cloud, deployment, extractedStemcell, registry, sshTunnelConfig, "fake-mbus-url")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-wait-running-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Waiting for 'fake-job-name'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "Waiting for 'fake-job-name': fake-wait-running-error",
			}))
		})
	})
})
