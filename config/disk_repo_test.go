package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("DiskRepo", func() {
	var (
		configService     DeploymentConfigService
		diskRepo          DiskRepo
		fs                *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		diskRepo = NewDiskRepo(configService, fakeUUIDGenerator)

	})

	Describe("Save", func() {
		It("saves the disk record using the config service", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
			record, err := diskRepo.Save("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(DiskRecord{
				ID:  "fake-guid-1",
				CID: "fake-cid",
			}))

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentConfig{
				Disks: []DiskRecord{
					{
						ID:  "fake-guid-1",
						CID: "fake-cid",
					},
				},
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})
	})

	Describe("Find", func() {
		It("finds existing disk records", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
			savedRecord, err := diskRepo.Save("fake-cid")
			Expect(err).ToNot(HaveOccurred())

			foundRecord, found, err := diskRepo.Find("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(foundRecord).To(Equal(savedRecord))
		})

		It("when the disk is not in the records, returns not found", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-guid-2"
			_, err := diskRepo.Save("other-cid")
			Expect(err).ToNot(HaveOccurred())

			_, found, err := diskRepo.Find("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("FindCurrent", func() {
		Context("when current disk exists", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := diskRepo.Save("fake-cid-1")
				Expect(err).ToNot(HaveOccurred())

				fakeUUIDGenerator.GeneratedUuid = "fake-guid-2"
				record, err := diskRepo.Save("fake-cid-2")
				Expect(err).ToNot(HaveOccurred())

				diskRepo.UpdateCurrent(record)
			})

			It("returns existing disk", func() {
				record, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(record).To(Equal(DiskRecord{
					ID:  "fake-guid-2",
					CID: "fake-cid-2",
				}))
			})
		})

		Context("when current disk does not exist", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := diskRepo.Save("fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns not found", func() {
				_, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when there are no disks", func() {
			It("returns not found", func() {
				_, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
