package vm_test

import (
	"errors"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk/fakes"
	fakebmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
)

var _ = Describe("DiskDeployer", func() {
	var (
		diskDeployer    DiskDeployer
		fakeDiskManager *fakebmdisk.FakeManager
		diskPool        bmmanifest.DiskPool
		cloud           *fakebmcloud.FakeCloud
		fakeStage       *fakebmlog.FakeStage
		fakeVM          *fakebmvm.FakeVM
		fakeDisk        *fakebmdisk.FakeDisk
		fakeDiskRepo    *fakebmconfig.FakeDiskRepo
	)

	BeforeEach(func() {
		cloud = fakebmcloud.NewFakeCloud()
		fakeVM = fakebmvm.NewFakeVM("fake-vm-cid")

		fakeDiskManagerFactory := fakebmdisk.NewFakeManagerFactory()
		fakeDiskManager = fakebmdisk.NewFakeManager()
		fakeDisk = fakebmdisk.NewFakeDisk("fake-new-disk-cid")
		fakeDiskManager.CreateDisk = fakeDisk
		fakeDiskManagerFactory.NewManagerManager = fakeDiskManager

		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeStage = fakebmlog.NewFakeStage()
		fakeDiskRepo = fakebmconfig.NewFakeDiskRepo()
		diskDeployer = NewDiskDeployer(
			fakeDiskManagerFactory,
			fakeDiskRepo,
			logger,
		)

		fakeDiskManager.SetFindCurrentBehavior(nil, false, nil)
		fakeVM.SetAttachDiskBehavior(fakeDisk, nil)
		newDiskRecord := bmconfig.DiskRecord{
			ID: "fake-new-disk-id",
		}
		fakeDiskRepo.SetFindBehavior("fake-new-disk-cid", newDiskRecord, true, nil)
	})

	Context("when the disk pool size is > 0", func() {
		BeforeEach(func() {
			diskPool = bmmanifest.DiskPool{
				Name:     "fake-persistent-disk-pool-name",
				DiskSize: 1024,
				RawCloudProperties: map[interface{}]interface{}{
					"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
				},
			}
		})

		Context("when primary disk already exists", func() {
			var existingDisk *fakebmdisk.FakeDisk

			BeforeEach(func() {
				existingDisk = fakebmdisk.NewFakeDisk("fake-existing-disk-cid")
				fakeDiskManager.SetFindCurrentBehavior(existingDisk, true, nil)
				fakeVM.SetAttachDiskBehavior(existingDisk, nil)
				existingDiskRecord := bmconfig.DiskRecord{
					ID: "fake-existing-disk-id",
				}
				fakeDiskRepo.SetFindBehavior("fake-existing-disk-cid", existingDiskRecord, true, nil)
			})

			It("does not create primary disk", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDiskManager.CreateInputs).To(BeEmpty())
			})

			Context("when disk does not need migration", func() {
				BeforeEach(func() {
					existingDisk.SetNeedsMigrationBehavior(false)
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

			Context("when disk needs migration", func() {
				var secondaryDisk *fakebmdisk.FakeDisk

				BeforeEach(func() {
					existingDisk.SetNeedsMigrationBehavior(true)

					secondaryDisk = fakebmdisk.NewFakeDisk("fake-secondary-disk-cid")
					fakeDiskManager.CreateDisk = secondaryDisk
					secondaryDiskRecord := bmconfig.DiskRecord{
						ID: "fake-secondary-disk-id",
					}

					fakeDiskRepo.SetFindBehavior("fake-secondary-disk-cid", secondaryDiskRecord, true, nil)
				})

				It("creates secondary disk", func() {
					err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeDiskManager.CreateInputs).To(Equal([]fakebmdisk.CreateInput{
						{
							DiskPool:   diskPool,
							InstanceID: "fake-vm-cid",
						},
					}))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Creating disk",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					}))
				})

				It("attaches secondary disk", func() {
					err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeVM.AttachDiskInputs).To(Equal([]fakebmvm.AttachDiskInput{
						{Disk: existingDisk},
						{Disk: secondaryDisk},
					}))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					}))
				})

				It("migrates from primary to secondary disk", func() {
					err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeVM.MigrateDiskCalledTimes).To(Equal(1))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					}))
				})

				It("detaches primary disk", func() {
					err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeVM.DetachDiskInputs).To(Equal([]fakebmvm.DetachDiskInput{
						{Disk: existingDisk},
					}))

					Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
						Name: "Detaching disk 'fake-existing-disk-cid'",
						States: []bmeventlog.EventState{
							bmeventlog.Started,
							bmeventlog.Finished,
						},
					}))
				})

				It("promotes secondary disk as primary", func() {
					err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).NotTo(HaveOccurred())

					// existing disk must be current until after migration
					Expect(fakeDiskRepo.UpdateCurrentInputs).To(Equal([]fakebmconfig.DiskRepoUpdateCurrentInput{
						//						{ DiskID: "fake-existing-disk-id" },
						{DiskID: "fake-secondary-disk-id"},
					}))
				})

				Context("when disk creation fails", func() {
					BeforeEach(func() {
						fakeDiskManager.CreateErr = errors.New("fake-create-disk-error")
					})

					It("returns error and leaves the existing disk attached", func() {
						err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-create-disk-error"))
						Expect(fakeVM.DetachDiskInputs).To(Equal([]fakebmvm.DetachDiskInput{}))
					})
				})

				Context("when attaching the new disk fails", func() {
					BeforeEach(func() {
						fakeVM.SetAttachDiskBehavior(secondaryDisk, errors.New("fake-attach-disk-error"))
					})

					It("returns error and leaves the existing disk attached", func() {
						err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-attach-disk-error"))
						Expect(fakeVM.DetachDiskInputs).To(Equal([]fakebmvm.DetachDiskInput{}))

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Failed,
							},
							FailMessage: "fake-attach-disk-error",
						}))
					})
				})

				Context("when detaching the new disk fails", func() {
					BeforeEach(func() {
						fakeVM.SetDetachDiskBehavior(existingDisk, errors.New("fake-detach-disk-error"))
					})

					It("returns error", func() {
						err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-detach-disk-error"))

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Detaching disk 'fake-existing-disk-cid'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Failed,
							},
							FailMessage: "fake-detach-disk-error",
						}))
					})
				})

				Context("when migration to the new disk fails", func() {
					BeforeEach(func() {
						fakeVM.MigrateDiskErr = errors.New("fake-migrate-disk-error")
					})

					It("returns error and leaves the existing disk attached", func() {
						err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migrate-disk-error"))
						Expect(fakeVM.DetachDiskInputs).To(Equal([]fakebmvm.DetachDiskInput{}))

						Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
							Name: "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'",
							States: []bmeventlog.EventState{
								bmeventlog.Started,
								bmeventlog.Failed,
							},
							FailMessage: "fake-migrate-disk-error",
						}))
					})
				})
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

			It("sets the new disk as current", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDiskRepo.UpdateCurrentInputs).To(Equal([]fakebmconfig.DiskRepoUpdateCurrentInput{
					{DiskID: "fake-new-disk-id"},
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

		It("attaches the primary disk", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.AttachDiskInputs).To(Equal([]fakebmvm.AttachDiskInput{
				{
					Disk: fakeDisk,
				},
			}))
		})

		It("logs attaching primary disk event", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Attaching disk 'fake-new-disk-cid' to VM 'fake-vm-cid'",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		It("removes unused disks", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeDiskManager.DeleteUnusedCalledTimes).To(Equal(1))
		})

		Context("when removing unused disk fails", func() {
			BeforeEach(func() {
				fakeDiskManager.DeleteUnusedErr = errors.New("fake-delete-error")
			})

			It("returns an error", func() {
				err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))
			})
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
				fakeVM.SetAttachDiskBehavior(fakeDisk, errors.New("fake-attach-disk-error"))
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
					Name: "Attaching disk 'fake-new-disk-cid' to VM 'fake-vm-cid'",
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
			diskPool = bmmanifest.DiskPool{}
		})

		It("does not create a persistent disk", func() {
			err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeDiskManager.CreateInputs).To(BeEmpty())
		})
	})
})
