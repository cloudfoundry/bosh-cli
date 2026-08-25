package vm_test

import (
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	"github.com/cloudfoundry/bosh-cli/v7/config/configfakes"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/disk/diskfakes"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	. "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	"github.com/cloudfoundry/bosh-cli/v7/deployment/vm/vmfakes"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("DiskDeployer", func() {
	var (
		diskDeployer           DiskDeployer
		fakeDiskManager        *diskfakes.FakeManager
		diskPool               bideplmanifest.DiskPool
		cloud                  *cloudfakes.FakeCloud
		fakeStage              *testui.Stage
		fakeVM                 *vmfakes.FakeVM
		fakeDisk               *diskfakes.FakeDisk
		fakeDiskRepo           *configfakes.FakeDiskRepo
		fakeDiskManagerFactory *diskfakes.FakeManagerFactory
		logger                 boshlog.Logger
	)

	BeforeEach(func() {
		cloud = &cloudfakes.FakeCloud{}
		fakeVM = &vmfakes.FakeVM{CIDStub: func() string { return "fake-vm-cid" }}

		fakeDiskManagerFactory = &diskfakes.FakeManagerFactory{
			NewManagerStub: func(cloud bicloud.Cloud) bidisk.Manager {
				return fakeDiskManager
			},
		}
		fakeDiskManager = &diskfakes.FakeManager{
			CreateStub: func(pool bideplmanifest.DiskPool, s string) (bidisk.Disk, error) {
				return fakeDisk, nil
			},
		}
		fakeDisk = &diskfakes.FakeDisk{
			CIDStub: func() string { return "fake-new-disk-cid" },
		}

		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeStage = &testui.Stage{}
		fakeDiskRepo = &configfakes.FakeDiskRepo{}
		diskDeployer = NewDiskDeployer(
			fakeDiskManagerFactory,
			fakeDiskRepo,
			logger,
			false,
		)

		fakeDiskManager.FindCurrentReturns([]bidisk.Disk{}, nil)
		fakeVM.AttachDiskReturns(nil)
		newDiskRecord := biconfig.DiskRecord{
			ID: "fake-new-disk-id",
		}
		fakeDiskRepo.FindReturnsOnCall(0, newDiskRecord, true, nil)
	})

	Context("when the disk pool size is > 0", func() {
		BeforeEach(func() {
			diskPool = bideplmanifest.DiskPool{
				Name:     "fake-persistent-disk-pool-name",
				DiskSize: 1024,
				CloudProperties: biproperty.Map{
					"fake-disk-pool-cloud-property-key": "fake-disk-pool-cloud-property-value",
				},
			}
		})

		Context("when primary disk already exists", func() {
			var existingDisk *diskfakes.FakeDisk
			var attachDiskResponses = map[bidisk.Disk]error{}

			BeforeEach(func() {
				existingDisk = &diskfakes.FakeDisk{CIDStub: func() string {
					return "fake-existing-disk-cid"
				}}
				fakeDiskManager.FindCurrentReturns([]bidisk.Disk{existingDisk}, nil)
				attachDiskResponses[existingDisk] = nil
				fakeVM.AttachDiskStub = func(disk bidisk.Disk) error {
					return attachDiskResponses[disk]
				}
				existingDiskRecord := biconfig.DiskRecord{
					ID: "fake-existing-disk-id",
				}
				fakeDiskRepo.FindReturnsOnCall(1, existingDiskRecord, true, nil)
			})

			It("does not create primary disk", func() {
				disks, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDiskManager.CreateCallCount()).To(Equal(0))
				Expect(disks).To(Equal([]bidisk.Disk{existingDisk}))
			})

			Context("when disk does not need migration", func() {
				BeforeEach(func() {
					existingDisk.NeedsMigrationReturns(false, nil)
				})

				It("does not log the create disk event", func() {
					disks, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(disks).To(Equal([]bidisk.Disk{existingDisk}))

					Expect(fakeStage.PerformCalls).ToNot(ContainElement(&testui.PerformCall{
						Name: "Creating disk",
					}))
				})
			})

			Context("when disk is forced to be recreated", func() {
				var secondaryDisk *diskfakes.FakeDisk

				BeforeEach(func() {
					diskDeployer = NewDiskDeployer(
						fakeDiskManagerFactory,
						fakeDiskRepo,
						logger,
						true,
					)
					existingDisk.NeedsMigrationReturns(false, nil)

					secondaryDisk = &diskfakes.FakeDisk{CIDStub: func() string {
						return "fake-secondary-disk-cid"
					}}
					fakeDiskManager.CreateReturns(secondaryDisk, nil)
					secondaryDiskRecord := biconfig.DiskRecord{
						ID: "fake-secondary-disk-id",
					}

					fakeDiskRepo.FindReturnsOnCall(0, secondaryDiskRecord, true, nil)
				})

				It("creates secondary disk", func() {
					disks, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(disks).To(Equal([]bidisk.Disk{secondaryDisk}))

					pool, diskCid := fakeDiskManager.CreateArgsForCall(0)
					Expect(pool).To(Equal(diskPool))
					Expect(diskCid).To(Equal("fake-vm-cid"))

					Expect(fakeStage.PerformCalls[1]).To(Equal(&testui.PerformCall{
						Name: "Creating disk",
					}))
				})

				It("attaches secondary disk", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeVM.AttachDiskArgsForCall(0)).To(Equal(existingDisk))
					Expect(fakeVM.AttachDiskArgsForCall(1)).To(Equal(secondaryDisk))

					Expect(fakeStage.PerformCalls[2]).To(Equal(&testui.PerformCall{
						Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'",
					}))
				})

				It("migrates from primary to secondary disk", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeVM.MigrateDiskCallCount()).To(Equal(1))

					Expect(fakeStage.PerformCalls[3]).To(Equal(&testui.PerformCall{
						Name: "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'",
					}))
				})

				It("detaches primary disk", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeVM.DetachDiskArgsForCall(0)).To(Equal(existingDisk))

					Expect(fakeStage.PerformCalls[4]).To(Equal(&testui.PerformCall{
						Name: "Detaching disk 'fake-existing-disk-cid'",
					}))
				})

				It("promotes secondary disk as primary", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).NotTo(HaveOccurred())

					// existing disk must be current until after migration
					Expect(fakeDiskRepo.UpdateCurrentArgsForCall(0)).To(Equal("fake-secondary-disk-id"))
				})

				Context("when disk creation fails", func() {
					BeforeEach(func() {
						fakeDiskManager.CreateReturns(nil, bosherr.Error("fake-create-disk-error"))
					})

					It("returns error and leaves the existing disk attached", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-create-disk-error"))
						Expect(fakeVM.DetachDiskCallCount()).To(Equal(0))
					})
				})

				Context("when attaching the new disk fails", func() {
					var (
						attachError = bosherr.Error("fake-attach-disk-error")
					)

					BeforeEach(func() {
						attachDiskResponses[secondaryDisk] = attachError
					})

					It("returns error and leaves the existing disk attached", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-attach-disk-error"))
						Expect(fakeVM.DetachDiskCallCount()).To(Equal(0))

						Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
							{Name: "Attaching disk 'fake-existing-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Creating disk"},
							{
								Name:  "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'",
								Error: attachError,
							},
						}))
					})
				})

				Context("when detaching the new disk fails", func() {
					var (
						detachError = bosherr.Error("fake-detach-disk-error")
					)

					BeforeEach(func() {
						fakeVM.DetachDiskStub = func(disk bidisk.Disk) error {
							switch disk {
							case existingDisk:
								return detachError
							default:
								return nil
							}
						}
					})

					It("returns error", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-detach-disk-error"))

						Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
							{Name: "Attaching disk 'fake-existing-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Creating disk"},
							{Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'"},
							{
								Name:  "Detaching disk 'fake-existing-disk-cid'",
								Error: detachError,
							},
						}))
					})
				})

				Context("when migration to the new disk fails", func() {
					var (
						migrateError = bosherr.Error("fake-migrate-disk-error")
					)

					BeforeEach(func() {
						fakeVM.MigrateDiskReturns(migrateError)
					})

					It("returns error and leaves the existing disk attached", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migrate-disk-error"))
						Expect(fakeVM.DetachDiskCallCount()).To(Equal(0))

						Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
							{Name: "Attaching disk 'fake-existing-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Creating disk"},
							{Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'"},
							{
								Name:  "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'",
								Error: migrateError,
							},
						}))
					})
				})
			})

			Context("when disk needs migration", func() {
				var secondaryDisk *diskfakes.FakeDisk

				BeforeEach(func() {
					existingDisk.NeedsMigrationReturns(true, nil)

					secondaryDisk = &diskfakes.FakeDisk{CIDStub: func() string {
						return "fake-secondary-disk-cid"
					}}
					fakeDiskManager.CreateReturns(secondaryDisk, nil)
					secondaryDiskRecord := biconfig.DiskRecord{
						ID: "fake-secondary-disk-id",
					}

					fakeDiskRepo.FindReturnsOnCall(0, secondaryDiskRecord, true, nil)
				})

				It("creates secondary disk", func() {
					disks, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(disks).To(Equal([]bidisk.Disk{secondaryDisk}))

					pool, diskCid := fakeDiskManager.CreateArgsForCall(0)
					Expect(pool).To(Equal(diskPool))
					Expect(diskCid).To(Equal("fake-vm-cid"))

					Expect(fakeStage.PerformCalls[1]).To(Equal(&testui.PerformCall{
						Name: "Creating disk",
					}))
				})

				It("attaches secondary disk", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeVM.AttachDiskArgsForCall(0)).To(Equal(existingDisk))
					Expect(fakeVM.AttachDiskArgsForCall(1)).To(Equal(secondaryDisk))

					Expect(fakeStage.PerformCalls[2]).To(Equal(&testui.PerformCall{
						Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'",
					}))
				})

				It("migrates from primary to secondary disk", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).ToNot(HaveOccurred())
					Expect(fakeVM.MigrateDiskCallCount()).To(Equal(1))

					Expect(fakeStage.PerformCalls[3]).To(Equal(&testui.PerformCall{
						Name: "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'",
					}))
				})

				It("detaches primary disk", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeVM.DetachDiskArgsForCall(0)).To(Equal(existingDisk))

					Expect(fakeStage.PerformCalls[4]).To(Equal(&testui.PerformCall{
						Name: "Detaching disk 'fake-existing-disk-cid'",
					}))
				})

				It("promotes secondary disk as primary", func() {
					_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
					Expect(err).NotTo(HaveOccurred())

					// existing disk must be current until after migration
					Expect(fakeDiskRepo.UpdateCurrentArgsForCall(0)).To(Equal("fake-secondary-disk-id"))
				})

				Context("when disk creation fails", func() {
					BeforeEach(func() {
						fakeDiskManager.CreateReturns(nil, bosherr.Error("fake-create-disk-error"))
					})

					It("returns error and leaves the existing disk attached", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-create-disk-error"))
						Expect(fakeVM.DetachDiskCallCount()).To(Equal(0))
					})
				})

				Context("when attaching the new disk fails", func() {
					var (
						attachError = bosherr.Error("fake-attach-disk-error")
					)

					BeforeEach(func() {
						fakeVM.AttachDiskStub = func(disk bidisk.Disk) error {
							switch disk {
							case secondaryDisk:
								return attachError
							default:
								return nil
							}
						}
					})

					It("returns error and leaves the existing disk attached", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-attach-disk-error"))
						Expect(fakeVM.DetachDiskCallCount()).To(Equal(0))

						Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
							{Name: "Attaching disk 'fake-existing-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Creating disk"},
							{
								Name:  "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'",
								Error: attachError,
							},
						}))
					})
				})

				Context("when detaching the new disk fails", func() {
					var (
						detachError = bosherr.Error("fake-detach-disk-error")
					)

					BeforeEach(func() {
						fakeVM.DetachDiskReturns(detachError)
					})

					It("returns error", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-detach-disk-error"))

						Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
							{Name: "Attaching disk 'fake-existing-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Creating disk"},
							{Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'"},
							{
								Name:  "Detaching disk 'fake-existing-disk-cid'",
								Error: detachError,
							},
						}))
					})
				})

				Context("when migration to the new disk fails", func() {
					var (
						migrateError = bosherr.Error("fake-migrate-disk-error")
					)

					BeforeEach(func() {
						fakeVM.MigrateDiskReturns(migrateError)
					})

					It("returns error and leaves the existing disk attached", func() {
						_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-migrate-disk-error"))
						Expect(fakeVM.DetachDiskCallCount()).To(Equal(0))

						Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
							{Name: "Attaching disk 'fake-existing-disk-cid' to VM 'fake-vm-cid'"},
							{Name: "Creating disk"},
							{Name: "Attaching disk 'fake-secondary-disk-cid' to VM 'fake-vm-cid'"},
							{
								Name:  "Migrating disk content from 'fake-existing-disk-cid' to 'fake-secondary-disk-cid'",
								Error: migrateError,
							},
						}))
					})
				})
			})
		})

		Context("when disk does not exist", func() {
			It("creates a persistent disk", func() {
				disks, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(disks).To(Equal([]bidisk.Disk{fakeDisk}))

				pool, diskCid := fakeDiskManager.CreateArgsForCall(0)
				Expect(pool).To(Equal(diskPool))
				Expect(diskCid).To(Equal("fake-vm-cid"))
			})

			It("sets the new disk as current", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeDiskRepo.UpdateCurrentArgsForCall(0)).To(Equal("fake-new-disk-id"))
			})

			It("logs the create disk event", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0]).To(Equal(&testui.PerformCall{
					Name: "Creating disk",
				}))
			})
		})

		It("attaches the primary disk", func() {
			_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVM.AttachDiskArgsForCall(0)).To(Equal(fakeDisk))
		})

		It("logs attaching primary disk event", func() {
			_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
				{Name: "Creating disk"},
				{Name: "Attaching disk 'fake-new-disk-cid' to VM 'fake-vm-cid'"},
			}))
		})

		It("removes unused disks", func() {
			_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeDiskManager.DeleteUnusedCallCount()).To(Equal(1))
		})

		Context("when removing unused disk fails", func() {
			BeforeEach(func() {
				fakeDiskManager.DeleteUnusedReturns(bosherr.Error("fake-delete-error"))
			})

			It("returns an error", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))
			})
		})

		Context("when creating the persistent disk fails", func() {
			var (
				createDiskError = bosherr.Error("fake-create-disk-error")
			)

			BeforeEach(func() {
				fakeDiskManager.CreateReturns(nil, createDiskError)
			})

			It("return an error", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-disk-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
					{
						Name:  "Creating disk",
						Error: createDiskError,
					},
				}))
			})
		})

		Context("when attaching the persistent disk fails", func() {
			var (
				attachDiskError = bosherr.Error("fake-attach-disk-error")
			)

			BeforeEach(func() {
				fakeVM.AttachDiskReturns(attachDiskError)
			})

			It("return an error", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-attach-disk-error"))
			})

			It("logs start and failed events to the eventLogger", func() {
				_, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.PerformCalls).To(Equal([]*testui.PerformCall{
					{Name: "Creating disk"},
					{
						Name:  "Attaching disk 'fake-new-disk-cid' to VM 'fake-vm-cid'",
						Error: attachDiskError,
					},
				}))
			})
		})
	})

	Context("when the disk pool size is 0", func() {
		BeforeEach(func() {
			diskPool = bideplmanifest.DiskPool{}
		})

		It("does not create a persistent disk", func() {
			disks, err := diskDeployer.Deploy(diskPool, cloud, fakeVM, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(disks).To(Equal([]bidisk.Disk{}))

			Expect(fakeDiskManager.CreateCallCount()).To(Equal(0))
		})
	})
})
