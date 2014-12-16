package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
)

var _ = Describe("Disk", func() {
	var (
		disk                Disk
		diskCloudProperties map[string]interface{}
		fakeCloud           *fakebmcloud.FakeCloud
		diskRepo            bmconfig.DiskRepo
	)

	BeforeEach(func() {
		diskCloudProperties = map[string]interface{}{
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
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		fakeUUIDGenerator := &fakeuuid.FakeGenerator{}
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
				newDiskCloudProperties := map[string]interface{}{
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
	})
})
