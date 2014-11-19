package disk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
)

var _ = Describe("Disk", func() {
	var (
		disk                Disk
		diskCloudProperties map[string]interface{}
	)

	BeforeEach(func() {
		diskCloudProperties = map[string]interface{}{
			"fake-cloud-property-key": "fake-cloud-property-value",
		}

		disk = NewDisk("fake-disk-cid", 1024, diskCloudProperties)
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
})
