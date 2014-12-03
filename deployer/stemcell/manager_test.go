package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
)

var _ = Describe("Manager", func() {
	var (
		stemcellRepo        bmconfig.StemcellRepo
		fakeUUIDGenerator   *fakeuuid.FakeGenerator
		manager             Manager
		fs                  *fakesys.FakeFileSystem
		reader              *fakebmstemcell.FakeStemcellReader
		fakeCloud           *fakebmcloud.FakeCloud
		fakeStage           *fakebmlog.FakeStage
		eventLogger         *fakebmlog.FakeEventLogger
		stemcellTarballPath string
		tempExtractionDir   string

		expectedExtractedStemcell ExtractedStemcell
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		reader = fakebmstemcell.NewFakeReader()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fs, logger)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		fakeUUIDGenerator.GeneratedUuid = "fake-stemcell-id-1"
		stemcellRepo = bmconfig.NewStemcellRepo(configService, fakeUUIDGenerator)
		eventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		eventLogger.SetNewStageBehavior(fakeStage)
		fakeCloud = fakebmcloud.NewFakeCloud()
		manager = NewManager(stemcellRepo, fakeCloud, eventLogger)
		stemcellTarballPath = "/stemcell/tarball/path"
		tempExtractionDir = "/path/to/dest"
		fs.TempDirDir = tempExtractionDir

		expectedExtractedStemcell = NewExtractedStemcell(
			Manifest{
				Name:      "fake-stemcell-name",
				Version:   "fake-stemcell-version",
				ImagePath: "fake-image-path",
				RawCloudProperties: map[interface{}]interface{}{
					"fake-prop-key": "fake-prop-value",
				},
			},
			ApplySpec{},
			tempExtractionDir,
			fs,
		)
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedExtractedStemcell, nil)
	})

	Describe("Upload", func() {
		var (
			expectedCloudStemcell CloudStemcell
		)

		BeforeEach(func() {
			fakeCloud.CreateStemcellCID = "fake-stemcell-cid"
			stemcellRecord := bmconfig.StemcellRecord{
				CID:     "fake-stemcell-cid",
				Name:    "fake-stemcell-name",
				Version: "fake-stemcell-version",
			}
			expectedCloudStemcell = NewCloudStemcell(stemcellRecord, stemcellRepo, fakeCloud)
		})

		It("starts a new event logger stage", func() {
			_, err := manager.Upload(expectedExtractedStemcell)
			Expect(err).ToNot(HaveOccurred())

			Expect(eventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
				{
					Name: "uploading stemcell",
				},
			}))

			Expect(fakeStage.Started).To(BeTrue())
			Expect(fakeStage.Finished).To(BeTrue())
		})

		It("uploads the stemcell to the infrastructure and returns the cid", func() {
			cloudStemcell, err := manager.Upload(expectedExtractedStemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloudStemcell).To(Equal(expectedCloudStemcell))

			Expect(fakeCloud.CreateStemcellInputs).To(Equal([]fakebmcloud.CreateStemcellInput{
				{
					CloudProperties: map[string]interface{}{
						"fake-prop-key": "fake-prop-value",
					},
					ImagePath: "fake-image-path",
				},
			}))
		})

		It("saves the stemcell record in the stemcellRepo", func() {
			cloudStemcell, err := manager.Upload(expectedExtractedStemcell)
			Expect(err).ToNot(HaveOccurred())
			Expect(cloudStemcell).To(Equal(expectedCloudStemcell))

			stemcellRecords, err := stemcellRepo.All()
			Expect(stemcellRecords).To(Equal([]bmconfig.StemcellRecord{
				{
					ID:      "fake-stemcell-id-1",
					Name:    "fake-stemcell-name",
					Version: "fake-stemcell-version",
					CID:     "fake-stemcell-cid",
				},
			}))
		})

		It("logs uploading start and stop events to the eventLogger", func() {
			_, err := manager.Upload(expectedExtractedStemcell)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Uploading",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		It("when the upload fails, logs uploading start and failure events to the eventLogger", func() {
			fakeCloud.CreateStemcellErr = errors.New("fake-create-error")
			_, err := manager.Upload(expectedExtractedStemcell)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-create-error"))

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Uploading",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Failed,
				},
				FailMessage: "creating stemcell (fake-stemcell-name fake-stemcell-version): fake-create-error",
			}))
		})

		It("when the stemcellRepo save fails, logs uploading start and failure events to the eventLogger", func() {
			fs.WriteToFileError = errors.New("fake-save-error")
			_, err := manager.Upload(expectedExtractedStemcell)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-save-error"))
			Expect(fakeStage.Steps).To(HaveLen(1))
			uploadStep := fakeStage.Steps[0]
			Expect(uploadStep.FailMessage).To(ContainSubstring("fake-save-error"))
			Expect(uploadStep.States).To(Equal([]bmeventlog.EventState{
				bmeventlog.Started,
				bmeventlog.Failed,
			}))
			Expect(uploadStep.Name).To(Equal("Uploading"))
		})

		Context("when the stemcell record exists in the stemcellRepo (having been previously uploaded)", func() {
			var (
				foundStemcellRecord bmconfig.StemcellRecord
			)

			BeforeEach(func() {
				var err error
				foundStemcellRecord, err = stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-existing-cid")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the existing cloud stemcell", func() {
				stemcell, err := manager.Upload(expectedExtractedStemcell)
				Expect(err).ToNot(HaveOccurred())
				foundStemcell := NewCloudStemcell(foundStemcellRecord, stemcellRepo, fakeCloud)
				Expect(stemcell).To(Equal(foundStemcell))
			})

			It("does not re-upload the stemcell to the infrastructure", func() {
				_, err := manager.Upload(expectedExtractedStemcell)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeCloud.CreateStemcellInputs).To(HaveLen(0))
			})

			It("logs skipping uploading events to the eventLogger", func() {
				_, err := manager.Upload(expectedExtractedStemcell)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Uploading",
					States: []bmeventlog.EventState{
						bmeventlog.Skipped,
					},
					SkipMessage: "Stemcell already uploaded",
				}))

				Expect(fakeStage.Finished).To(BeTrue())
			})
		})
	})

	Describe("FindCurrent", func() {
		Context("when stemcell already exists in stemcell repo", func() {
			BeforeEach(func() {
				stemcellRecord, err := stemcellRepo.Save("fake-stemcell-name", "fake-stemcell-version", "fake-existing-stemcell-cid")
				Expect(err).ToNot(HaveOccurred())

				err = stemcellRepo.UpdateCurrent(stemcellRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the existing stemcell", func() {
				stemcell, found, err := manager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(stemcell.CID()).To(Equal("fake-existing-stemcell-cid"))
			})
		})

		Context("when stemcell does not exists in stemcell repo", func() {
			It("returns false", func() {
				_, found, err := manager.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when reading stemcell repo fails", func() {
			BeforeEach(func() {
				fs.WriteFileString("/fake/path", "{}")
				fs.ReadFileError = errors.New("fake-read-error")
			})

			It("returns an error", func() {
				_, found, err := manager.FindCurrent()
				Expect(found).To(BeFalse())
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
			fakeUUIDGenerator.GeneratedUuid = "fake-stemcell-id-1"
			firstStemcellRecord, err := stemcellRepo.Save("fake-stemcell-name-1", "fake-stemcell-version-1", "fake-stemcell-cid-1")
			Expect(err).ToNot(HaveOccurred())
			firstStemcell = NewCloudStemcell(firstStemcellRecord, stemcellRepo, fakeCloud)

			fakeUUIDGenerator.GeneratedUuid = "fake-stemcell-id-2"
			_, err = stemcellRepo.Save("fake-stemcell-name-2", "fake-stemcell-version-2", "fake-stemcell-cid-2")
			Expect(err).ToNot(HaveOccurred())
			err = stemcellRepo.UpdateCurrent("fake-stemcell-id-2")
			Expect(err).ToNot(HaveOccurred())

			fakeUUIDGenerator.GeneratedUuid = "fake-stemcell-id-3"
			secondStemcellRecord, err := stemcellRepo.Save("fake-stemcell-name-3", "fake-stemcell-version-3", "fake-stemcell-cid-3")
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
})
