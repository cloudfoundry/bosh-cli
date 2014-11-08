package disk_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
)

var _ = Describe("Manager", func() {
	Describe("Create", func() {
		var (
			manager       Manager
			fakeCloud     *fakebmcloud.FakeCloud
			configService bmconfig.DeploymentConfigService
			diskPool      bmdepl.DiskPool
		)

		BeforeEach(func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			fs := fakesys.NewFakeFileSystem()
			configService = bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
			managerFactory := NewManagerFactory(configService, logger)
			fakeCloud = fakebmcloud.NewFakeCloud()
			manager = managerFactory.NewManager(fakeCloud)
			diskPool = bmdepl.DiskPool{
				Name: "fake-disk-pool-name",
				Size: 1024,
				RawCloudProperties: map[interface{}]interface{}{
					"fake-cloud-property-key": "fake-cloud-property-value",
				},
			}
		})

		Context("when disk already exists in deployment config", func() {
			BeforeEach(func() {
				configService.Save(bmconfig.DeploymentConfig{
					DiskCID: "fake-existing-disk-cid",
				})
			})

			It("returns the existing disk", func() {
				disk, err := manager.Create(diskPool, "fake-instance-id")
				Expect(err).ToNot(HaveOccurred())
				Expect(disk.CID()).To(Equal("fake-existing-disk-cid"))
			})
		})

		Context("when creating disk succeeds", func() {
			BeforeEach(func() {
				fakeCloud.CreateDiskCID = "fake-disk-cid"
			})

			It("returns a disk", func() {
				disk, err := manager.Create(diskPool, "fake-instance-id")
				Expect(err).ToNot(HaveOccurred())
				Expect(disk.CID()).To(Equal("fake-disk-cid"))
			})

			It("saves the disk record using the config service", func() {
				_, err := manager.Create(diskPool, "fake-instance-id")
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				expectedConfig := bmconfig.DeploymentConfig{
					DiskCID: "fake-disk-cid",
				}
				Expect(deploymentConfig).To(Equal(expectedConfig))
			})
		})

		Context("when creating disk fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateDiskErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(diskPool, "fake-instance-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})
	})
})
