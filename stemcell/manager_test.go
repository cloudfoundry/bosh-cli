package stemcell_test

import (
	"errors"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	"github.com/cloudfoundry/bosh-cli/v7/config/configfakes"
	. "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	fakebistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell/stemcellfakes"
	fakebiui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("Manager", func() {
	var (
		stemcellRepo        biconfig.StemcellRepo
		fakeUUIDGenerator   *fakeuuid.FakeGenerator
		manager             Manager
		fs                  *fakesys.FakeFileSystem
		reader              *fakebistemcell.FakeStemcellReader
		fakeCloud           *cloudfakes.FakeCloud
		fakeStage           *fakebiui.FakeStage
		stemcellTarballPath string
		tempExtractionDir   string

		expectedExtractedStemcell ExtractedStemcell
		stemcellApiVersion        = 2
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		reader = fakebistemcell.NewFakeReader()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		deploymentStateService := biconfig.NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, filepath.Join("/", "fake", "path"))
		fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-1"
		stemcellRepo = biconfig.NewStemcellRepo(deploymentStateService, fakeUUIDGenerator)
		fakeStage = fakebiui.NewFakeStage()
		fakeCloud = &cloudfakes.FakeCloud{}
		manager = NewManager(stemcellRepo, fakeCloud)
		stemcellTarballPath = filepath.Join("/", "stemcell", "tarball", "path")
		tempExtractionDir = filepath.Join("/", "path", "to", "dest")
		fs.TempDirDir = tempExtractionDir

		expectedExtractedStemcell = NewExtractedStemcell(
			Manifest{
				Name:    "fake-stemcell-name",
				Version: "fake-stemcell-version",
				CloudProperties: biproperty.Map{
					"fake-prop-key": "fake-prop-value",
				},
			},
			tempExtractionDir,
			nil,
			fs,
		)
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedExtractedStemcell, nil)
	})

	Describe("Upload", func() {
		var (
			expectedCloudStemcell CloudStemcell
		)

		BeforeEach(func() {
			fakeCloud.CreateStemcellReturns("fake-stemcell-cid", nil)
			stemcellRecord := biconfig.StemcellRecord{
				CID:     "fake-stemcell-cid",
				Name:    "fake-stemcell-name",
				Version: "fake-stemcell-version",
			}
			expectedCloudStemcell = NewCloudStemcell(stemcellRecord, stemcellRepo, fakeCloud)
		})

		It("uploads the stemcell to the infrastructure and returns the cid", func() {
			cloudStemcell, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloudStemcell).To(Equal(expectedCloudStemcell))

			Expect(fakeCloud.CreateStemcellCallCount()).To(Equal(1))
			imagePath, cloudProperties := fakeCloud.CreateStemcellArgsForCall(0)
			Expect(imagePath).To(Equal(filepath.Join(tempExtractionDir, "image")))
			Expect(cloudProperties).To(Equal(biproperty.Map{"fake-prop-key": "fake-prop-value"}))
		})

		It("saves the stemcell record in the stemcellRepo", func() {
			cloudStemcell, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloudStemcell).To(Equal(expectedCloudStemcell))

			stemcellRecords, err := stemcellRepo.All()
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcellRecords).To(Equal([]biconfig.StemcellRecord{
				{
					ID:      "fake-stemcell-id-1",
					Name:    "fake-stemcell-name",
					Version: "fake-stemcell-version",
					CID:     "fake-stemcell-cid",
				},
			}))
		})

		It("prints uploading ui stage", func() {
			_, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Uploading stemcell 'fake-stemcell-name/fake-stemcell-version'"},
			}))
		})

		It("when the upload fails, prints failed uploading ui stage", func() {
			fakeCloud.CreateStemcellReturns("", errors.New("fake-create-error"))
			_, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-error"))

			Expect(fakeStage.PerformCalls[0].Name).To(Equal("Uploading stemcell 'fake-stemcell-name/fake-stemcell-version'"))
			Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
			Expect(fakeStage.PerformCalls[0].Error.Error()).To(Equal("creating stemcell (fake-stemcell-name fake-stemcell-version): fake-create-error"))
		})

		It("when the stemcellRepo save fails, logs uploading start and failure events to the eventLogger", func() {
			fs.WriteFileError = errors.New("fake-save-error")
			_, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-save-error"))

			Expect(fakeStage.PerformCalls[0].Name).To(Equal("Uploading stemcell 'fake-stemcell-name/fake-stemcell-version'"))
			Expect(fakeStage.PerformCalls[0].Error).To(HaveOccurred())
			Expect(fakeStage.PerformCalls[0].Error.Error()).To(MatchRegexp("Finding existing stemcell record in repo: .*fake-save-error.*"))
		})

		Context("when the stemcell record exists in the stemcellRepo (having been previously uploaded)", func() {
			var (
				foundStemcellRecord biconfig.StemcellRecord
			)

			BeforeEach(func() {
				var err error
				foundStemcellRecord, err = stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-existing-cid", stemcellApiVersion)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the existing cloud stemcell", func() {
				stemcell, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
				Expect(err).ToNot(HaveOccurred())
				foundStemcell := NewCloudStemcell(foundStemcellRecord, stemcellRepo, fakeCloud)
				Expect(stemcell).To(Equal(foundStemcell))
			})

			It("does not re-upload the stemcell to the infrastructure", func() {
				_, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeCloud.CreateStemcellCallCount()).To(Equal(0))
			})

			It("logs skipping uploading events to the eventLogger", func() {
				_, err := manager.Upload(expectedExtractedStemcell, fakeStage, false)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.PerformCalls[0].Name).To(Equal("Uploading stemcell 'fake-stemcell-name/fake-stemcell-version'"))
				Expect(fakeStage.PerformCalls[0].SkipError).To(HaveOccurred())
				Expect(fakeStage.PerformCalls[0].SkipError.Error()).To(MatchRegexp("Stemcell already uploaded: Found stemcell: .*fake-existing-cid.*"))
			})

			Context("when fix is requested", func() {
				BeforeEach(func() {
					err := stemcellRepo.UpdateCurrent(foundStemcellRecord.ID)
					Expect(err).ToNot(HaveOccurred())
				})

				It("re-uploads the stemcell to the infrastructure", func() {
					_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCloud.CreateStemcellCallCount()).To(Equal(1))
					imagePath, cloudProperties := fakeCloud.CreateStemcellArgsForCall(0)
					Expect(imagePath).To(Equal(filepath.Join(tempExtractionDir, "image")))
					Expect(cloudProperties).To(Equal(biproperty.Map{"fake-prop-key": "fake-prop-value"}))
				})

				It("returns the newly created stemcell, not the stale one", func() {
					cloudStemcell, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(cloudStemcell.CID()).To(Equal("fake-stemcell-cid"))
				})

				It("replaces the stale record rather than failing on a duplicate", func() {
					_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())

					stemcellRecords, err := stemcellRepo.All()
					Expect(err).ToNot(HaveOccurred())
					Expect(stemcellRecords).To(HaveLen(1))
					Expect(stemcellRecords[0].CID).To(Equal("fake-stemcell-cid"))
				})

				It("never leaves CurrentStemcellID empty, which would strand delete-env on CPI api version 1", func() {
					_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())

					currentRecord, found, err := stemcellRepo.FindCurrent()
					Expect(err).ToNot(HaveOccurred())
					Expect(found).To(BeTrue(), "CurrentStemcellID must never be left empty")
					Expect(currentRecord.CID).To(Equal("fake-stemcell-cid"))
				})

				It("does not report live stemcells as unused, which on AWS would deregister the AMI", func() {
					_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())

					unused, err := manager.FindUnused()
					Expect(err).ToNot(HaveOccurred())
					Expect(unused).To(BeEmpty(), "a blanked CurrentStemcellID would mark every stemcell unused")
				})

				It("leaves the replaced image in the cloud as the rollback target", func() {
					_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCloud.DeleteStemcellCallCount()).To(Equal(0))
				})

				It("does not skip the upload stage", func() {
					_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeStage.PerformCalls[0].SkipError).ToNot(HaveOccurred())
				})

				Context("when the upload fails partway, as a long transfer may", func() {
					BeforeEach(func() {
						fakeCloud.CreateStemcellReturns("", errors.New("fake-create-error"))
					})

					It("leaves the existing record intact", func() {
						_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
						Expect(err).To(HaveOccurred())

						records, err := stemcellRepo.All()
						Expect(err).ToNot(HaveOccurred())
						Expect(records).To(HaveLen(1))
						Expect(records[0].CID).To(Equal("fake-existing-cid"))
					})

					It("leaves CurrentStemcellID intact", func() {
						_, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
						Expect(err).To(HaveOccurred())

						currentRecord, found, err := stemcellRepo.FindCurrent()
						Expect(err).ToNot(HaveOccurred())
						Expect(found).To(BeTrue(), "a failed upload must not blank CurrentStemcellID")
						Expect(currentRecord.CID).To(Equal("fake-existing-cid"))
					})
				})
			})
		})

		// A fake repo: the fs-backed one fails at Find before Save is reached.
		Context("when saving the stemcell record fails", func() {
			var (
				fakeRepo    *configfakes.FakeStemcellRepo
				fakeManager Manager
			)

			BeforeEach(func() {
				fakeRepo = &configfakes.FakeStemcellRepo{}
				fakeRepo.FindReturns(biconfig.StemcellRecord{}, false, nil)
				fakeRepo.SaveReturns(biconfig.StemcellRecord{}, errors.New("fake-save-error"))
				fakeManager = NewManager(fakeRepo, fakeCloud)
			})

			It("deletes the orphaned stemcell from the cloud", func() {
				_, err := fakeManager.Upload(expectedExtractedStemcell, fakeStage, false)
				Expect(err).To(HaveOccurred())

				Expect(fakeCloud.DeleteStemcellCallCount()).To(Equal(1))
				Expect(fakeCloud.DeleteStemcellArgsForCall(0)).To(Equal("fake-stemcell-cid"))
			})

			It("reports the save failure, not the cleanup result", func() {
				_, err := fakeManager.Upload(expectedExtractedStemcell, fakeStage, false)
				Expect(err.Error()).To(ContainSubstring("fake-save-error"))
			})

			Context("when deleting the orphaned stemcell also fails", func() {
				BeforeEach(func() {
					fakeCloud.DeleteStemcellReturns(errors.New("fake-delete-error"))
				})

				It("still reports the save failure, mentioning the leak", func() {
					_, err := fakeManager.Upload(expectedExtractedStemcell, fakeStage, false)
					Expect(err.Error()).To(ContainSubstring("fake-save-error"))
					Expect(err.Error()).To(ContainSubstring("fake-delete-error"))
				})
			})

			Context("when the returned CID was already tracked in the repo", func() {
				BeforeEach(func() {
					fakeRepo.FindReturns(biconfig.StemcellRecord{CID: "fake-stemcell-cid"}, true, nil)
					fakeRepo.SaveOrUpdateReturns(biconfig.StemcellRecord{}, errors.New("fake-save-error"))
				})

				It("does not delete the pre-existing stemcell from the cloud", func() {
					_, err := fakeManager.Upload(expectedExtractedStemcell, fakeStage, true)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-save-error"))
					Expect(fakeCloud.DeleteStemcellCallCount()).To(Equal(0))
				})
			})
		})

		Context("when no stemcell record exists and fix is requested", func() {
			It("uploads the stemcell as usual", func() {
				cloudStemcell, err := manager.Upload(expectedExtractedStemcell, fakeStage, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(cloudStemcell).To(Equal(expectedCloudStemcell))
				Expect(fakeCloud.CreateStemcellCallCount()).To(Equal(1))
			})
		})
	})

	Describe("FindCurrent", func() {
		Context("when stemcell already exists in stemcell repo", func() {
			BeforeEach(func() {
				stemcellRecord, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-existing-stemcell-cid", stemcellApiVersion)
				Expect(err).ToNot(HaveOccurred())

				err = stemcellRepo.UpdateCurrent(stemcellRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the existing stemcell", func() {
				stemcells, err := manager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcells).To(HaveLen(1))
				Expect(stemcells[0].CID()).To(Equal("fake-existing-stemcell-cid"))
			})
		})

		Context("when stemcell does not exists in stemcell repo", func() {
			It("returns false", func() {
				stemcells, err := manager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcells).To(BeEmpty())
			})
		})

		Context("when reading stemcell repo fails", func() {
			BeforeEach(func() {
				err := fs.WriteFileString("/fake/path", "{}")
				Expect(err).ToNot(HaveOccurred())
				fs.ReadFileError = errors.New("fake-read-error")
			})

			It("returns an error", func() {
				_, err := manager.FindCurrent()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-read-error"))
			})
		})
	})

	Describe("FindUnused", func() {
		var (
			firstStemcell  CloudStemcell
			secondStemcell CloudStemcell
		)

		BeforeEach(func() {
			fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-1"
			firstStemcellRecord, err := stemcellRepo.Save("fake-stemcell-name-1", "fake-stemcell-version-1", "fake-stemcell-cid-1", stemcellApiVersion)
			Expect(err).ToNot(HaveOccurred())
			firstStemcell = NewCloudStemcell(firstStemcellRecord, stemcellRepo, fakeCloud)

			fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-2"
			_, err = stemcellRepo.Save("fake-stemcell-name-2", "fake-stemcell-version-2", "fake-stemcell-cid-2", stemcellApiVersion)
			Expect(err).ToNot(HaveOccurred())
			err = stemcellRepo.UpdateCurrent("fake-stemcell-id-2")
			Expect(err).ToNot(HaveOccurred())

			fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-3"
			secondStemcellRecord, err := stemcellRepo.Save("fake-stemcell-name-3", "fake-stemcell-version-3", "fake-stemcell-cid-3", stemcellApiVersion)
			Expect(err).ToNot(HaveOccurred())
			secondStemcell = NewCloudStemcell(secondStemcellRecord, stemcellRepo, fakeCloud)
		})

		It("returns unused stemcells", func() {
			stemcells, err := manager.FindUnused()
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcells).To(Equal([]CloudStemcell{
				firstStemcell,
				secondStemcell,
			}))
		})
	})

	Describe("DeleteUnused", func() {
		var (
			secondStemcellRecord biconfig.StemcellRecord
		)
		BeforeEach(func() {
			fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-1"
			_, err := stemcellRepo.Save("fake-stemcell-name-1", "fake-stemcell-version-1", "fake-stemcell-cid-1", stemcellApiVersion)
			Expect(err).ToNot(HaveOccurred())

			fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-2"
			secondStemcellRecord, err = stemcellRepo.Save("fake-stemcell-name-2", "fake-stemcell-version-2", "fake-stemcell-cid-2", stemcellApiVersion)
			Expect(err).ToNot(HaveOccurred())
			err = stemcellRepo.UpdateCurrent(secondStemcellRecord.ID)
			Expect(err).ToNot(HaveOccurred())

			fakeUUIDGenerator.GeneratedUUID = "fake-stemcell-id-3"
			_, err = stemcellRepo.Save("fake-stemcell-name-3", "fake-stemcell-version-3", "fake-stemcell-cid-3", stemcellApiVersion)
			Expect(err).ToNot(HaveOccurred())
		})

		It("deletes unused stemcells", func() {
			err := manager.DeleteUnused(fakeStage)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCloud.DeleteStemcellCallCount()).To(Equal(2))
			Expect(fakeCloud.DeleteStemcellArgsForCall(0)).To(Equal("fake-stemcell-cid-1"))
			Expect(fakeCloud.DeleteStemcellArgsForCall(1)).To(Equal("fake-stemcell-cid-3"))

			Expect(fakeStage.PerformCalls).To(Equal([]*fakebiui.PerformCall{
				{Name: "Deleting unused stemcell 'fake-stemcell-cid-1'"},
				{Name: "Deleting unused stemcell 'fake-stemcell-cid-3'"},
			}))

			currentRecord, found, err := stemcellRepo.FindCurrent()
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(currentRecord).To(Equal(secondStemcellRecord))

			records, err := stemcellRepo.All()
			Expect(err).ToNot(HaveOccurred())
			Expect(records).To(Equal([]biconfig.StemcellRecord{
				secondStemcellRecord,
			}))
		})
	})
})
