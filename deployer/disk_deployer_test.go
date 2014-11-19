package deployer_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("DiskDeployer", func() {
	var (
		diskDeployer    DiskDeployer
		fakeDiskManager *fakebmdisk.FakeManager
		diskPool        bmdepl.DiskPool
		cloud           *fakebmcloud.FakeCloud
		fakeStage       *fakebmlog.FakeStage
		fakeVM          *fakebmvm.FakeVM
	)

	BeforeEach(func() {
		cloud = fakebmcloud.NewFakeCloud()
		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")

		fakeDiskManagerFactory := fakebmdisk.NewFakeManagerFactory()
		fakeDiskManager = fakebmdisk.NewFakeManager()
		fakeDiskManager.CreateDisk = bmdisk.NewDisk("fake-disk-cid")
		fakeDiskManagerFactory.NewManagerManager = fakeDiskManager

		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeStage = fakebmlog.NewFakeStage()
		diskDeployer = NewDiskDeployer(
			fakeDiskManagerFactory,
			logger,
		)

		fakeDiskManager.SetFindCurrentBehavior(nil, false, nil)
	})

	Context("when the disk pool size is > 0", func() {
		BeforeEach(func() {
			diskPool = bmdepl.DiskPool{
				Name: "fake-persistent-disk-pool-name",
				Size: 1024,
				RawCloudProperties: map[interface{}]interface{}{
					"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
				},
			}
		})

		Context("when disk already exists", func() {
			BeforeEach(func() {
				disk := bmdisk.NewDisk("fake-disk-cid")
				fakeDiskManager.SetFindCurrentBehavior(disk, true, nil)
			})

			It("does not create disk", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDiskManager.CreateInputs).To(BeEmpty())
			})

			It("does not log the create disk event", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).ToNot(ContainElement(&fakebmlog.FakeStep{
					Name: "Creating disk",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				}))
			})
		})

		Context("when disk does not exist", func() {
			It("creates a persistent disk", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDiskManager.CreateInputs).To(Equal([]fakebmdisk.CreateInput{
					{
						DiskPool:   diskPool,
						InstanceID: "fake-vm-cid",
					},
				}))
			})

			It("logs the create disk event", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Creating disk",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				}))
			})
		})

		It("attaches the persistent disk", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.AttachDiskInputs).To(Equal([]fakebmvm.AttachDiskInput{
				{
					Disk: bmdisk.NewDisk("fake-disk-cid"),
				},
			}))
		})

		It("logs attaching persistent disk event", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Attaching disk 'fake-disk-cid' to VM 'fake-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when creating the persistent disk fails", func() {
			BeforeEach(func() {
				fakeDiskManager.CreateErr = errors.New("fake-create-disk-error")
			})

			It("return an error", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-disk-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Creating disk",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-create-disk-error",
				}))
			})
		})

		Context("when attaching the persistent disk fails", func() {
			BeforeEach(func() {
				fakeVM.AttachDiskErr = errors.New("fake-attach-disk-error")
			})

			It("return an error", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-attach-disk-error"))
			})

			It("logs start and failed events to the eventLogger", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Attaching disk 'fake-disk-cid' to VM 'fake-vm-cid'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-attach-disk-error",
				}))
			})
		})
	})

	Context("when the disk pool size is 0", func() {
		BeforeEach(func() {
			diskPool = bmdepl.DiskPool{}
		})

		It("does not create a persistent disk", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDiskManager.CreateInputs).To(BeEmpty())
		})
	})
})
