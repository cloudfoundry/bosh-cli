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
		repo              DiskRepo
		fs                *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		repo = NewDiskRepo(configService, fakeUUIDGenerator)
	})

	Describe("Save", func() {
		It("saves the disk record using the config service", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
			record, err := repo.Save("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(DiskRecord{
				ID:  "fake-guid-1",
				CID: "fake-cid",
			}))

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentFile{
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
			savedRecord, err := repo.Save("fake-cid")
			Expect(err).ToNot(HaveOccurred())

			foundRecord, found, err := repo.Find("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(foundRecord).To(Equal(savedRecord))
		})

		It("when the disk is not in the records, returns not found", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-guid-2"
			_, err := repo.Save("other-cid")
			Expect(err).ToNot(HaveOccurred())

			_, found, err := repo.Find("fake-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("UpdateCurrent", func() {
		Context("when a disk record exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
				_, err := repo.Save("fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the disk record as current stemcell", func() {
				err := repo.UpdateCurrent("fake-uuid-1")
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig.CurrentDiskID).To(Equal("fake-uuid-1"))
			})
		})

		Context("when a disk record does not exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
				_, err := repo.Save("fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				err := repo.UpdateCurrent("fake-uuid-2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying disk record exists with id `fake-uuid-2'"))
			})
		})
	})

	Describe("FindCurrent", func() {
		Context("when current disk exists", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := repo.Save("fake-cid-1")
				Expect(err).ToNot(HaveOccurred())

				fakeUUIDGenerator.GeneratedUuid = "fake-guid-2"
				record, err := repo.Save("fake-cid-2")
				Expect(err).ToNot(HaveOccurred())

				repo.UpdateCurrent(record.ID)
			})

			It("returns existing disk", func() {
				record, found, err := repo.FindCurrent()
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
				_, err := repo.Save("fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns not found", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when there are no disks", func() {
			It("returns not found", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
