package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("DeploymentRecord", func() {
	var (
		deploymentRecord        DeploymentRecord
		deploymentConfigService DeploymentConfigService
		fakeFs                  *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		deploymentConfigService = bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fakeFs, logger)
		deploymentRecord = NewDeploymentRecord(deploymentConfigService, logger)
	})

	Describe("GetDisk", func() {
		Context("when disk record is present", func() {
			BeforeEach(func() {
				deploymentConfig := DeploymentConfig{
					DiskCID: "fake-disk-cid",
				}
				deploymentConfigService.Save(deploymentConfig)
			})

			It("returns disk record", func() {
				disk, found, err := deploymentRecord.Disk()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(disk).To(Equal(DiskRecord{
					CID: "fake-disk-cid",
				}))
			})
		})

		Context("when disk record is not present", func() {
			BeforeEach(func() {
				deploymentConfigService.Save(DeploymentConfig{})
			})

			It("returns false", func() {
				_, found, err := deploymentRecord.Disk()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when loading the deployment config fails", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString("/fake/path", "")
			})

			It("returns error", func() {
				_, found, err := deploymentRecord.Disk()
				Expect(err).To(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("UpdateDisk", func() {
		BeforeEach(func() {
			deploymentConfig := DeploymentConfig{
				DiskCID: "fake-disk-cid",
			}
			deploymentConfigService.Save(deploymentConfig)
		})

		It("updates disk record", func() {
			newRecord := DiskRecord{
				CID: "new-fake-disk-cid",
			}
			err := deploymentRecord.UpdateDisk(newRecord)
			Expect(err).ToNot(HaveOccurred())

			disk, found, err := deploymentRecord.Disk()
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(disk).To(Equal(newRecord))

			deploymentConfig, err := deploymentConfigService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentConfig.DiskCID).To(Equal("new-fake-disk-cid"))
		})
	})
})
