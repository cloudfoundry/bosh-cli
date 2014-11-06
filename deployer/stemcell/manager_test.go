package stemcell_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
)

var _ = Describe("Manager", func() {
	var (
		repo                *fakebmstemcell.FakeRepo
		manager             Manager
		fs                  *fakesys.FakeFileSystem
		reader              *fakebmstemcell.FakeStemcellReader
		fakeCloud           *fakebmcloud.FakeCloud
		eventLogger         *fakebmlog.FakeEventLogger
		stemcellTarballPath string
		tempExtractionDir   string

		expectedStemcell Stemcell
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		reader = fakebmstemcell.NewFakeReader()
		repo = fakebmstemcell.NewFakeRepo()
		eventLogger = fakebmlog.NewFakeEventLogger()
		fakeCloud = fakebmcloud.NewFakeCloud()
		managerFactory := NewManagerFactory(fs, reader, repo, eventLogger)
		manager = managerFactory.NewManager(fakeCloud)
		stemcellTarballPath = "/stemcell/tarball/path"
		tempExtractionDir = "/path/to/dest"
		fs.TempDirDir = tempExtractionDir

		expectedStemcell = Stemcell{
			Manifest: Manifest{
				Name:      "fake-stemcell-name",
				ImagePath: "fake-image-path",
				RawCloudProperties: map[interface{}]interface{}{
					"fake-prop-key": "fake-prop-value",
				},
			},
		}
		reader.SetReadBehavior(stemcellTarballPath, tempExtractionDir, expectedStemcell, nil)

		// no existing stemcell found
		repo.SetFindBehavior(expectedStemcell.Manifest, CID(""), false, nil)

		fakeCloud.CreateStemcellCID = "fake-stemcell-cid"

		repo.SetSaveBehavior(expectedStemcell.Manifest, "fake-stemcell-cid", nil)
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
		Expect(cid).To(Equal(CID("fake-stemcell-cid")))

		Expect(fakeCloud.CreateStemcellInputs).To(Equal([]fakebmcloud.CreateStemcellInput{
			{
				CloudProperties: map[string]interface{}{
					"fake-prop-key": "fake-prop-value",
				},
				ImagePath: "fake-image-path",
			},
		}))
	})

	It("saves the stemcell record in the repo", func() {
		_, cid, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(cid).To(Equal(CID("fake-stemcell-cid")))

		Expect(repo.SaveInputs).To(Equal(
			[]fakebmstemcell.SaveInput{
				fakebmstemcell.SaveInput{
					StemcellManifest: expectedStemcell.Manifest,
					CID:              "fake-stemcell-cid",
				},
			},
		))
	})

	It("logs unpacking start and stop events to the eventLogger", func() {
		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		expectedStartEvent := bmeventlog.Event{
			Stage: "Uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Started,
		}

		expectedFinishEvent := bmeventlog.Event{
			Stage: "Uploading stemcell",
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
			Stage: "Uploading stemcell",
			Total: 2,
			Task:  "Unpacking",
			Index: 1,
			State: bmeventlog.Started,
		}

		expectedFailedEvent := bmeventlog.Event{
			Stage:   "Uploading stemcell",
			Total:   2,
			Task:    "Unpacking",
			Index:   1,
			State:   bmeventlog.Failed,
			Message: "fake-read-error",
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(2))
	})

	It("logs uploading start and stop events to the eventLogger", func() {
		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).ToNot(HaveOccurred())

		expectedStartEvent := bmeventlog.Event{
			Stage: "Uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Started,
		}

		expectedFinishEvent := bmeventlog.Event{
			Stage: "Uploading stemcell",
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
		fakeCloud.CreateStemcellErr = errors.New("fake-create-error")
		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-create-error"))

		expectedStartEvent := bmeventlog.Event{
			Stage: "Uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Started,
		}

		expectedFailedEvent := bmeventlog.Event{
			Stage:   "Uploading stemcell",
			Total:   2,
			Task:    "Uploading",
			Index:   2,
			State:   bmeventlog.Failed,
			Message: "fake-create-error",
		}

		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
		Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFailedEvent))
		Expect(eventLogger.LoggedEvents).To(HaveLen(4))
	})

	It("when the repo save fails, logs uploading start and failure events to the eventLogger", func() {
		repo.SetSaveBehavior(
			expectedStemcell.Manifest,
			"fake-stemcell-cid",
			errors.New("fake-save-error"),
		)

		_, _, err := manager.Upload(stemcellTarballPath)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake-save-error"))

		expectedStartEvent := bmeventlog.Event{
			Stage: "Uploading stemcell",
			Total: 2,
			Task:  "Uploading",
			Index: 2,
			State: bmeventlog.Started,
		}

		expectedFailedEvent := bmeventlog.Event{
			Stage:   "Uploading stemcell",
			Total:   2,
			Task:    "Uploading",
			Index:   2,
			State:   bmeventlog.Failed,
			Message: "fake-save-error",
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
			Expect(fakeCloud.CreateStemcellInputs).To(HaveLen(0))
		})

		It("logs skipping uploading events to the eventLogger", func() {
			_, _, err := manager.Upload(stemcellTarballPath)
			Expect(err).ToNot(HaveOccurred())

			expectedStartEvent := bmeventlog.Event{
				Stage: "Uploading stemcell",
				Total: 2,
				Task:  "Unpacking",
				Index: 1,
				State: bmeventlog.Started,
			}

			expectedFinishEvent := bmeventlog.Event{
				Stage: "Uploading stemcell",
				Total: 2,
				Task:  "Unpacking",
				Index: 1,
				State: bmeventlog.Finished,
			}

			expectedSkipEvent := bmeventlog.Event{
				Stage:   "Uploading stemcell",
				Total:   2,
				Task:    "Uploading",
				Index:   2,
				State:   bmeventlog.Skipped,
				Message: "Stemcell already uploaded",
			}

			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedStartEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedFinishEvent))
			Expect(eventLogger.LoggedEvents).To(ContainElement(expectedSkipEvent))
			Expect(eventLogger.LoggedEvents).To(HaveLen(3))
		})
	})
})
