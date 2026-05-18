package config_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/config"
)

var _ = Describe("DiskRepo", func() {
	var (
		deploymentStateService DeploymentStateService
		vmRepo                 VMRepo
		repo                   DiskRepo
		fs                     *fakesys.FakeFileSystem
		fakeUUIDGenerator      *fakeuuid.FakeGenerator
		cloudProperties        biproperty.Map
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		deploymentStateService = NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, "/fake/path")
		vmRepo = NewVMRepo(deploymentStateService, fakeUUIDGenerator)
		repo = NewDiskRepo(deploymentStateService, fakeUUIDGenerator)
		cloudProperties = biproperty.Map{
			"fake-cloud_property-key": "fake-cloud-property-value",
		}
	})

	Describe("Save", func() {
		It("saves the disk record using the config service", func() {
			record, err := repo.Save("fake-cid", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(DiskRecord{
				ID:              "fake-uuid-1",
				CID:             "fake-cid",
				Size:            1024,
				CloudProperties: cloudProperties,
			}))

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())

			Expect(deploymentState.Disks).To(ContainElement(DiskRecord{
				ID:              "fake-uuid-1",
				CID:             "fake-cid",
				Size:            1024,
				CloudProperties: cloudProperties,
			}))
		})
	})

	Describe("Find", func() {
		It("finds existing disk records", func() {
			savedRecord, err := repo.Save("fake-cid", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())

			foundRecord, found, err := repo.Find("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(foundRecord).To(Equal(savedRecord))
		})

		It("when the disk is not in the records, returns not found", func() {
			_, err := repo.Save("other-cid", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())

			_, found, err := repo.Find("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("UpdateCurrentForVM / FindCurrentForVM", func() {
		var vmCID = "vm-cid-1"

		BeforeEach(func() {
			_, err := vmRepo.Save("nats", 0, vmCID, "")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when a disk record exists with the same ID", func() {
			var recordID string

			BeforeEach(func() {
				record, err := repo.Save("fake-cid", 1024, cloudProperties)
				Expect(err).ToNot(HaveOccurred())
				recordID = record.ID
			})

			It("associates the disk with the VM", func() {
				err := repo.UpdateCurrentForVM(vmCID, recordID)
				Expect(err).ToNot(HaveOccurred())

				diskRecord, found, err := repo.FindCurrentForVM(vmCID)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(diskRecord.ID).To(Equal(recordID))
				Expect(diskRecord.CID).To(Equal("fake-cid"))
			})
		})

		Context("when a disk record does not exist with the given ID", func() {
			It("returns an error", func() {
				err := repo.UpdateCurrentForVM(vmCID, "fake-unknown-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying disk record exists with id 'fake-unknown-id'"))
			})
		})
	})

	Describe("FindCurrentForVM", func() {
		var vmCID = "vm-cid-1"

		BeforeEach(func() {
			_, err := vmRepo.Save("nats", 0, vmCID, "")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when current disk exists for the VM", func() {
			var diskID2 string

			BeforeEach(func() {
				_, err := repo.Save("fake-cid-1", 1024, cloudProperties)
				Expect(err).ToNot(HaveOccurred())

				record, err := repo.Save("fake-cid-2", 1024, cloudProperties)
				Expect(err).ToNot(HaveOccurred())
				diskID2 = record.ID

				err = repo.UpdateCurrentForVM(vmCID, record.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the disk for that VM", func() {
				record, found, err := repo.FindCurrentForVM(vmCID)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(record.ID).To(Equal(diskID2))
				Expect(record.CID).To(Equal("fake-cid-2"))
			})
		})

		Context("when no current disk exists for the VM", func() {
			It("returns not found", func() {
				_, found, err := repo.FindCurrentForVM(vmCID)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when querying a VM CID not in state", func() {
			It("returns not found", func() {
				_, found, err := repo.FindCurrentForVM("unknown-vm")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("ClearCurrentForVM", func() {
		var vmCID = "vm-cid-1"

		BeforeEach(func() {
			_, err := vmRepo.Save("nats", 0, vmCID, "")
			Expect(err).ToNot(HaveOccurred())
			record, err := repo.Save("fake-cid", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())
			err = repo.UpdateCurrentForVM(vmCID, record.ID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("clears the disk association for the VM", func() {
			err := repo.ClearCurrentForVM(vmCID)
			Expect(err).ToNot(HaveOccurred())

			_, found, err := repo.FindCurrentForVM(vmCID)
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("FindUnused", func() {
		var vmCID = "vm-cid-1"

		BeforeEach(func() {
			_, err := vmRepo.Save("nats", 0, vmCID, "")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns disks not referenced by any VM", func() {
			usedDisk, err := repo.Save("used-cid", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())
			unusedDisk, err := repo.Save("unused-cid", 2048, cloudProperties)
			Expect(err).ToNot(HaveOccurred())

			err = repo.UpdateCurrentForVM(vmCID, usedDisk.ID)
			Expect(err).ToNot(HaveOccurred())

			unused, err := repo.FindUnused()
			Expect(err).ToNot(HaveOccurred())
			Expect(unused).To(ConsistOf(unusedDisk))
		})

		It("returns all disks when no VM has a disk", func() {
			disk1, err := repo.Save("cid-1", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())
			disk2, err := repo.Save("cid-2", 2048, cloudProperties)
			Expect(err).ToNot(HaveOccurred())

			unused, err := repo.FindUnused()
			Expect(err).ToNot(HaveOccurred())
			Expect(unused).To(ConsistOf(disk1, disk2))
		})
	})

	Describe("All", func() {
		var (
			firstDisk  DiskRecord
			secondDisk DiskRecord
		)

		BeforeEach(func() {
			var err error
			firstDisk, err = repo.Save("fake-cid-1", 1024, cloudProperties)
			Expect(err).ToNot(HaveOccurred())

			secondDisk, err = repo.Save("fake-cid-2", 2048, cloudProperties)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns all disks", func() {
			disks, err := repo.All()
			Expect(err).ToNot(HaveOccurred())
			Expect(disks).To(Equal([]DiskRecord{
				firstDisk,
				secondDisk,
			}))
		})
	})

	Describe("Delete", func() {
		var (
			firstDisk  DiskRecord
			secondDisk DiskRecord
			vmCID      = "vm-cid-1"
		)

		BeforeEach(func() {
			_, err := vmRepo.Save("nats", 0, vmCID, "")
			Expect(err).ToNot(HaveOccurred())

			var e error
			firstDisk, e = repo.Save("fake-cid-1", 1024, cloudProperties)
			Expect(e).ToNot(HaveOccurred())

			secondDisk, e = repo.Save("fake-cid-2", 2048, cloudProperties)
			Expect(e).ToNot(HaveOccurred())
		})

		It("removes the disk record from the repo", func() {
			err := repo.Delete(firstDisk)
			Expect(err).ToNot(HaveOccurred())

			disks, err := repo.All()
			Expect(err).ToNot(HaveOccurred())
			Expect(disks).To(Equal([]DiskRecord{secondDisk}))
		})

		Context("when the disk to be deleted is the current disk for a VM", func() {
			BeforeEach(func() {
				err := repo.UpdateCurrentForVM(vmCID, firstDisk.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("clears the disk association from the VMRecord", func() {
				err := repo.Delete(firstDisk)
				Expect(err).ToNot(HaveOccurred())

				disks, err := repo.All()
				Expect(err).ToNot(HaveOccurred())
				Expect(disks).To(Equal([]DiskRecord{secondDisk}))

				_, found, err := repo.FindCurrentForVM(vmCID)
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
