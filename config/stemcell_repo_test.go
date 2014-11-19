package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("StemcellRepo", func() {
	var (
		repo              StemcellRepo
		configService     DeploymentConfigService
		fs                *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		fakeUUIDGenerator.GeneratedUuid = "fake-uuid"
		repo = NewStemcellRepo(configService, fakeUUIDGenerator)
	})

	Describe("Save", func() {
		It("saves the stemcell record using the config service", func() {
			_, err := repo.Save("fake-name", "fake-version", "fake-cid")
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentFile{
				Stemcells: []StemcellRecord{
					{
						ID:      "fake-uuid",
						Name:    "fake-name",
						Version: "fake-version",
						CID:     "fake-cid",
					},
				},
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})

		It("return the stemcell record with a new uuid", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
			record, err := repo.Save("fake-name", "fake-version-1", "fake-cid-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(StemcellRecord{
				ID:      "fake-uuid-1",
				Name:    "fake-name",
				Version: "fake-version-1",
				CID:     "fake-cid-1",
			}))

			fakeUUIDGenerator.GeneratedUuid = "fake-uuid-2"
			record, err = repo.Save("fake-name", "fake-version-2", "fake-cid-2")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(StemcellRecord{
				ID:      "fake-uuid-2",
				Name:    "fake-name",
				Version: "fake-version-2",
				CID:     "fake-cid-2",
			}))
		})

		Context("when a stemcell record with the same name and version exists", func() {
			BeforeEach(func() {
				_, err := repo.Save("fake-name", "fake-version", "fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := repo.Save("fake-name", "fake-version", "fake-cid-2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate name/version"))
			})
		})

		Context("when there stemcell record with the same cid exists", func() {
			BeforeEach(func() {
				_, err := repo.Save("fake-name", "fake-version", "fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := repo.Save("fake-name-2", "fake-version-2", "fake-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate cid"))
			})
		})
	})

	Describe("Find", func() {
		Context("when a stemcell record with the same name and version exists", func() {
			BeforeEach(func() {
				_, err := repo.Save("fake-name", "fake-version", "fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("finds existing stemcell records", func() {
				foundStemcellRecord, found, err := repo.Find("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(foundStemcellRecord).To(Equal(StemcellRecord{
					ID:      "fake-uuid",
					Name:    "fake-name",
					Version: "fake-version",
					CID:     "fake-cid",
				}))
			})
		})

		Context("when a stemcell record with the same name and version does not exist", func() {
			It("finds existing stemcell records", func() {
				_, found, err := repo.Find("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("UpdateCurrent", func() {
		Context("when a stemcell record exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
				_, err := repo.Save("fake-name", "fake-version", "fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the stemcell record as current stemcell", func() {
				err := repo.UpdateCurrent("fake-uuid-1")
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig.CurrentStemcellID).To(Equal("fake-uuid-1"))
			})
		})

		Context("when a stemcell record does not exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
				_, err := repo.Save("fake-name", "fake-version", "fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				err := repo.UpdateCurrent("fake-uuid-2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying stemcell record exists with id `fake-uuid-2'"))
			})
		})
	})

	Describe("FindCurrent", func() {
		Context("when current stemcell exists", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := repo.Save("fake-name", "fake-version-1", "fake-cid-1")
				Expect(err).ToNot(HaveOccurred())

				fakeUUIDGenerator.GeneratedUuid = "fake-guid-2"
				record, err := repo.Save("fake-name", "fake-version-2", "fake-cid-2")
				Expect(err).ToNot(HaveOccurred())

				repo.UpdateCurrent(record.ID)
			})

			It("returns existing stemcell", func() {
				record, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(record).To(Equal(StemcellRecord{
					ID:      "fake-guid-2",
					Name:    "fake-name",
					Version: "fake-version-2",
					CID:     "fake-cid-2",
				}))
			})
		})

		Context("when current stemcell does not exist", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := repo.Save("fake-name", "fake-version", "fake-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns not found", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when there are no stemcells", func() {
			It("returns not found", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
