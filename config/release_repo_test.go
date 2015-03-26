package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/config"
)

var _ = Describe("ReleaseRepo", rootDesc)

func rootDesc() {
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
			fakeUUIDGenerator.GeneratedUUID = "fake-uuid-1"
			record, err := repo.Save("fake-name", "fake-version-1")
			Expect(err).ToNot(HaveOccurred())
			Expect(record).To(Equal(ReleaseRecord{
				ID:      "fake-uuid-1",
				Name:    "fake-name",
				Version: "fake-version-1",
			}))

			fakeUUIDGenerator.GeneratedUUID = "fake-uuid-2"
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
				fakeUUIDGenerator.GeneratedUUID = "fake-uuid-1"
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
			})

			It("saves the release record as current release", func() {
				ids := []string{"fake-uuid-1"}
				err := repo.UpdateCurrent(ids)
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig.CurrentReleaseIDs).To(Equal([]string{"fake-uuid-1"}))
			})

			It("can save multiple current releases", func() {
				fakeUUIDGenerator.GeneratedUUID = "fake-uuid-2"
				_, err := repo.Save("fake-name-x", "fake-version")
				Expect(err).ToNot(HaveOccurred())

				ids := []string{"fake-uuid-1", "fake-uuid-2"}
				err = repo.UpdateCurrent(ids)
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(deploymentConfig.CurrentReleaseIDs).To(Equal([]string{"fake-uuid-1", "fake-uuid-2"}))
			})
		})

		Context("when a release record does not exists with the same ID", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUUID = "fake-uuid-1"
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				err := repo.UpdateCurrent([]string{"fake-uuid-2"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying release record exists with id 'fake-uuid-2'"))
			})

			It("returns an error if any of the releas IDs are not found", func() {
				err := repo.UpdateCurrent([]string{"fake-uuid-1", "not-saved-id"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying release record exists with id 'not-saved-id'"))
			})
		})
	})

	Describe("FindCurrent", func() {
		Context("when a current release exists", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUUID = "fake-guid-a"
				recordA, err := repo.Save("fake-name-a", "fake-version-a")
				Expect(err).ToNot(HaveOccurred())

				fakeUUIDGenerator.GeneratedUUID = "fake-guid-b"
				_, err = repo.Save("fake-name-b", "fake-version-b")
				Expect(err).ToNot(HaveOccurred())

				fakeUUIDGenerator.GeneratedUUID = "fake-guid-c"
				recordC, err := repo.Save("fake-name-c", "fake-version-c")
				Expect(err).ToNot(HaveOccurred())

				repo.UpdateCurrent([]string{recordA.ID, recordC.ID})
			})

			It("returns existing release", func() {
				records, found, err := repo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(records).To(Equal([]ReleaseRecord{
					{
						ID:      "fake-guid-a",
						Name:    "fake-name-a",
						Version: "fake-version-a",
					},
					{
						ID:      "fake-guid-c",
						Name:    "fake-name-c",
						Version: "fake-version-c",
					},
				}))
			})
		})

		Context("when a current release does not exist", func() {
			BeforeEach(func() {
				fakeUUIDGenerator.GeneratedUUID = "fake-guid-1"
				_, err := repo.Save("fake-name", "fake-version")
				Expect(err).ToNot(HaveOccurred())
				deploymentConfig, err := configService.Load()
				Expect(err).ToNot(HaveOccurred())

				deploymentConfig.CurrentReleaseIDs = []string{"fake-guid-1", "guid-not-saved"}
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
}
