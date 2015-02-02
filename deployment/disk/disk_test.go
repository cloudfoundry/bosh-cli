package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
)

var _ = Describe("Disk", func() {
	var (
		disk                Disk
		diskCloudProperties bmproperty.Map
		fakeCloud           *fakebmcloud.FakeCloud
		diskRepo            bmconfig.DiskRepo
		fakeUUIDGenerator   *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		diskCloudProperties = bmproperty.Map{
			"fake-cloud-property-key": "fake-cloud-property-value",
		}
		fakeCloud = fakebmcloud.NewFakeCloud()

		diskRecord := bmconfig.DiskRecord{
			CID:             "fake-disk-cid",
			Size:            1024,
			CloudProperties: diskCloudProperties,
		}

		fs := fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, fakeUUIDGenerator, logger)
		diskRepo = bmconfig.NewDiskRepo(configService, fakeUUIDGenerator)

		disk = NewDisk(diskRecord, fakeCloud, diskRepo)
	})

	Describe("NeedsMigration", func() {
		Context("when size is different", func() {
			It("returns true", func() {
				needsMigration := disk.NeedsMigration(2048, diskCloudProperties)
				Expect(needsMigration).To(BeTrue())
			})
		})

		Context("when cloud properties are different", func() {
			It("returns true", func() {
				newDiskCloudProperties := bmproperty.Map{
					"fake-cloud-property-key": "new-fake-cloud-property-value",
				}

				needsMigration := disk.NeedsMigration(1024, newDiskCloudProperties)
				Expect(needsMigration).To(BeTrue())
			})
		})

		Context("when cloud properties are nil", func() {
			It("returns true", func() {
				needsMigration := disk.NeedsMigration(1024, nil)
				Expect(needsMigration).To(BeTrue())
			})
		})

		Context("when size and cloud properties are the same", func() {
			It("returns false", func() {
				needsMigration := disk.NeedsMigration(1024, diskCloudProperties)
				Expect(needsMigration).To(BeFalse())
			})
		})
	})

	Describe("Delete", func() {
		It("deletes disk from cloud", func() {
			err := disk.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeCloud.DeleteDiskInputs).To(Equal([]fakebmcloud.DeleteDiskInput{
				{
					DiskCID: "fake-disk-cid",
				},
			}))
		})

		It("deletes disk from repo", func() {
			_, err := diskRepo.Save("fake-disk-cid", 1024, diskCloudProperties)
			Expect(err).ToNot(HaveOccurred())

			err = disk.Delete()
			Expect(err).ToNot(HaveOccurred())
			diskRecords, err := diskRepo.All()
			Expect(diskRecords).To(BeEmpty())
		})

		Context("when deleted disk is the current disk", func() {
			BeforeEach(func() {
				diskRecord, err := diskRepo.Save("fake-disk-cid", 1024, diskCloudProperties)
				Expect(err).ToNot(HaveOccurred())

				err = diskRepo.UpdateCurrent(diskRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("clears current disk in the disk repo", func() {
				err := disk.Delete()
				Expect(err).ToNot(HaveOccurred())

				_, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when deleting disk in the cloud fails", func() {
			BeforeEach(func() {
				fakeCloud.DeleteDiskErr = errors.New("fake-delete-disk-error")
			})

			It("returns an error", func() {
				err := disk.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-disk-error"))
			})
		})

		Context("when deleting disk in the cloud fails with DiskNotFoundError", func() {
			var deleteErr = bmcloud.NewCPIError("delete_vm", bmcloud.CmdError{
				Type:    bmcloud.DiskNotFoundError,
				Message: "fake-disk-not-found-message",
			})

			BeforeEach(func() {
				diskRecord, err := diskRepo.Save("fake-disk-cid", 1024, diskCloudProperties)
				Expect(err).ToNot(HaveOccurred())

				err = diskRepo.UpdateCurrent(diskRecord.ID)
				Expect(err).ToNot(HaveOccurred())

				fakeCloud.DeleteDiskErr = deleteErr
			})

			It("deletes disk in the cloud", func() {
				err := disk.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deleteErr))

				Expect(fakeCloud.DeleteDiskInputs).To(Equal([]fakebmcloud.DeleteDiskInput{
					{
						DiskCID: "fake-disk-cid",
					},
				}))
			})

			It("deletes disk in the disk repo", func() {
				err := disk.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deleteErr))

				diskRecords, err := diskRepo.All()
				Expect(diskRecords).To(BeEmpty())
			})

			It("clears current disk in the disk repo", func() {
				err := disk.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(deleteErr))

				_, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
