package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"

	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/stemcell"
)

var _ = Describe("Manager", func() {
	var (
		repo                *fakebmstemcell.FakeRepo
		manager             Manager
		fs                  *fakesys.FakeFileSystem
		reader              *fakebmstemcell.FakeStemcellReader
		infrastructure      *fakebmstemcell.FakeInfrastructure
		eventLogger         *fakebmlog.FakeEventLogger
		stemcellTarballPath string
		tempExtractionDir   string

		expectedStemcell Stemcell
		expectedCID      CID
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		reader = fakebmstemcell.NewFakeReader()
		repo = fakebmstemcell.NewFakeRepo()
		eventLogger = fakebmlog.NewFakeEventLogger()
		infrastructure = fakebmstemcell.NewFakeInfrastructure()
		managerFactory := NewManagerFactory(fs, reader, repo, eventLogger)
		manager = managerFactory.NewManager(infrastructure)
		stemcellTarballPath = "/stemcell/tarball/path"
		tempExtractionDir = "/path/to/dest"
		fs.TempDirDir = tempExtractionDir

		expectedStemcell = Stemcell{
			Manifest: Manifest{
				Name: "fake-stemcell-name",
			},
		}
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedStemcell, nil)

		// no existing stemcell found
		repo.SetFindBehavior(expectedStemcell.Manifest, CID(""), false, nil)

		expectedCID = "fake-cid"
		infrastructure.SetCreateStemcellBehavior(expectedStemcell.Manifest, expectedCID, nil)

		repo.SetSaveBehavior(expectedStemcell.Manifest, expectedCID, nil)
	})

	It("cleans up the temp work dir", func() {
		Expect(fs.FileExists(tempExtractionDir)).To(Equal(false))

		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(fs.FileExists(tempExtractionDir)).To(Equal(false))
	})

	It("extracts and parses the stemcell manifest", func() {
		stemcell, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(stemcell).To(Equal(expectedStemcell))

		Expect(reader.ReadInputs).To(Equal(
			[]fakebmstemcell.ReadInput{
				fakebmstemcell.ReadInput{
					StemcellTarballPath: stemcellTarballPath,
					DestPath:            tempExtractionDir,
				},
			},
		))
	})

	It("checks that the stemcell has not already been uploaded", func() {
		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(repo.FindInputs).To(Equal(
			[]fakebmstemcell.FindInput{
				fakebmstemcell.FindInput{
					StemcellManifest: expectedStemcell.Manifest,
				},
			},
		))
	})

	It("uploads the stemcell to the infrastructure and returns the cid", func() {
		_, cid, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(cid).To(Equal(expectedCID))

		Expect(infrastructure.CreateInputs).To(Equal(
			[]fakebmstemcell.CreateInput{
				{
					StemcellManifest: expectedStemcell.Manifest,
				},
			},
		))
	})

	It("saves the stemcell record in the repo", func() {
		_, cid, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(cid).To(Equal(expectedCID))

		Expect(repo.SaveInputs).To(Equal(
			[]fakebmstemcell.SaveInput{
				fakebmstemcell.SaveInput{
					StemcellManifest: expectedStemcell.Manifest,
					CID:              expectedCID,
				},
			},
		))
	})

	It("logs unpacking start and stop events to the eventLogger", func() {
		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		expectedStartEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Started,
		}

		expectedFinishEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Finished,
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFinishEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(4))
	})

	It("when the read fails, logs unpacking start and failure events to the eventLogger", func() {
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedStemcell, errors.New("fake-read-error"))

		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-read-error"))

		expectedStartEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Started,
		}

		expectedFailedEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Failed,
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(2))
	})

	It("logs uploading start and stop events to the eventLogger", func() {
		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		expectedStartEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Started,
		}

		expectedFinishEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Finished,
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFinishEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(4))
	})

	It("when the upload fails, logs uploading start and failure events to the eventLogger", func() {
		infrastructure.SetCreateStemcellBehavior(
			expectedStemcell.Manifest,
			expectedCID,
			errors.New("fake-create-error"),
		)

		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-create-error"))

		expectedStartEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Started,
		}

		expectedFailedEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Failed,
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(4))
	})

	It("when the repo save fails, logs uploading start and failure events to the eventLogger", func() {
		repo.SetSaveBehavior(
			expectedStemcell.Manifest,
			expectedCID,
			errors.New("fake-save-error"),
		)

		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-save-error"))

		expectedStartEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Started,
		}

		expectedFailedEvent := bmeventlog.Event{
			Stage: "uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Failed,
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(4))
	})

	Context("when the stemcell record exists in the repo (having been previously uploaded)", func() {
		var (
			existingCID CID
		)

		BeforeEach(func() {
			existingCID = CID("fake-cid")
			repo.SetFindBehavior(expectedStemcell.Manifest, existingCID, true, nil)
		})

		It("extracts and parses the stemcell manifest", func() {
			stemcell, _, err := manager.Upload(stemcellTarballPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(stemcell).To(Equal(expectedStemcell))

			Expect(reader.ReadInputs).To(Equal(
				[]fakebmstemcell.ReadInput{
					fakebmstemcell.ReadInput{
						StemcellTarballPath: stemcellTarballPath,
						DestPath:            tempExtractionDir,
					},
				},
			))
		})

		It("returns the cid of the existing stemcell", func() {
			_, cid, err := manager.Upload(stemcellTarballPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(cid).To(Equal(existingCID))
		})

		It("does not re-upload the stemcell to the infrastructure", func() {
			_, _, err := manager.Upload(stemcellTarballPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(infrastructure.CreateInputs).To(HaveLen(0))
		})

		It("logs skipping uploading events to the eventLogger", func() {
			_, _, err := manager.Upload(stemcellTarballPath)
			Expect(err).ToNot(HaveOccurred())

			expectedStartEvent := bmeventlog.Event{
				Stage: "uploading stemcell",
				Total: 2,
				Task:  "Unpacking",
				Index: 1,
				State: bmeventlog.Started,
			}

			expectedFinishEvent := bmeventlog.Event{
				Stage: "uploading stemcell",
				Total: 2,
				Task:  "Unpacking",
				Index: 1,
				State: bmeventlog.Finished,
			}

			expectedSkipEvent := bmeventlog.Event{
				Stage:   "uploading stemcell",
				Total:   2,
				Task:    "Uploading",
				Index:   2,
				State:   bmeventlog.Skipped,
				Message: "stemcell already uploaded",
			}

			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFinishEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedSkipEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(3))
		})
	})
})
