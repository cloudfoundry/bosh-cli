package disk_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
)

var _ = Describe("Manager", func() {
	Describe("Create", func() {
		var (
			manager         Manager
			fakeCloud       *fakebmcloud.FakeCloud
			cloudProperties map[string]interface{}
		)

		BeforeEach(func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			managerFactory := NewManagerFactory(logger)
			fakeCloud = fakebmcloud.NewFakeCloud()
			manager = managerFactory.NewManager(fakeCloud)
			cloudProperties = map[string]interface{}{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}
		})

		Context("when creating disk succeeds", func() {
			BeforeEach(func() {
				fakeCloud.CreateDiskCID = "fake-disk-cid"
			})

			It("returns a disk", func() {
				disk, err := manager.Create(1024, cloudProperties, "fake-instance-id")
				Expect(err).ToNot(HaveOccurred())
				Expect(disk).To(Equal(Disk{
					CID: "fake-disk-cid",
				}))
			})
		})

		Context("when creating disk fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateDiskErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(1024, cloudProperties, "fake-instance-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})
	})
})
