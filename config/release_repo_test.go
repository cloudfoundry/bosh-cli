package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("ReleaseRepo", func() {
	var (
		repo              ReleaseRepo
		configService     DeploymentConfigService
		fs                *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		configService = NewFileSystemDeploymentConfigService("/fake/path", fs, fakeUUIDGenerator, logger)
		repo = NewReleaseRepo(configService, fakeUUIDGenerator)
	})

	Describe("Save", func() {
		It("saves the release record using the config service", func() {
			_, err := repo.Save("fake-name", "fake-version")
			Expect(err).ToNot(HaveOccurred())

			deploymentConfig, err := configService.Load()
			Expect(err).ToNot(HaveOccurred())

			expectedConfig := DeploymentFile{
				DirectorID: "fake-uuid-0",
				Releases: []ReleaseRecord{
					{
						ID:      "fake-uuid-1",
						Name:    "fake-name",
						Version: "fake-version",
					},
				},
			}
			Expect(deploymentConfig).To(Equal(expectedConfig))
		})

		It("return the release record with a new uuid", func() {
			fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
			record, err := repo.Save("fake-name", "fake-version-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(ReleaseRecord{
				ID:      "fake-uuid-1",
				Name:    "fake-name",
				Version: "fake-version-1",
			}))

			fakeUUIDGenerator.GeneratedUuid = "fake-uuid-2"
			record, err = repo.Save("fake-name", "fake-version-2")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(ReleaseRecord{
				ID:      "fake-uuid-2",
				Name:    "fake-name",
				Version: "fake-version-2",
			}))
		})

		Context("when a release record with the same name and version exists", func() {
			BeforeEach(func() {
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate name/version"))
			})
		})
	})

	Describe("Find", func() {
		Context("when a release record with the same name and version exists", func() {
			var (
				recordID string
			)

			BeforeEach(func() {
				record, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
				recordID = record.ID
			})

			It("finds existing release records", func() {
				foundRecord, found, err := repo.Find("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(foundRecord).To(Equal(ReleaseRecord{
					ID:      recordID,
					Name:    "fake-name",
					Version: "fake-version",
				}))
			})
		})

		Context("when a release record with the same name and version does not exist", func() {
			It("finds existing release records", func() {
				_, found, err := repo.Find("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})

	Describe("UpdateCurrent", func() {
		Context("when a release record exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the release record as current release", func() {
				err := repo.UpdateCurrent("fake-uuid-1")
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig.CurrentReleaseID).To(Equal("fake-uuid-1"))
			})
		})

		Context("when a release record does not exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-uuid-1"
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				err := repo.UpdateCurrent("fake-uuid-2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying release record exists with id 'fake-uuid-2'"))
			})
		})
	})

	Describe("FindCurrent", func() {
		Context("when a current release exists", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := repo.Save("fake-name", "fake-version-1")
				Expect(err).ToNot(HaveOccurred())

				fakeUUIDGenerator.GeneratedUuid = "fake-guid-2"
				record, err := repo.Save("fake-name", "fake-version-2")
				Expect(err).ToNot(HaveOccurred())

				repo.UpdateCurrent(record.ID)
			})

			It("returns existing release", func() {
				record, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(record).To(Equal(ReleaseRecord{
					ID:      "fake-guid-2",
					Name:    "fake-name",
					Version: "fake-version-2",
				}))
			})
		})

		Context("when current release does not exist", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUuid = "fake-guid-1"
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns not found", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when there are no releases recorded", func() {
			It("returns not found", func() {
				_, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
