package templatescompiler_test

import (
	. "github.com/cloudfoundry/bosh-init/templatescompiler"

	"bytes"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"
	"os"

	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"

	fakeboshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("RenderedJob", func() {
	var (
		outBuffer *bytes.Buffer
		errBuffer *bytes.Buffer
		logger    boshlog.Logger
		fs        *fakeboshsys.FakeFileSystem

		releaseJob bireljob.Job

		renderedJobPath string

		renderedJob RenderedJob
	)

	BeforeEach(func() {
		outBuffer = bytes.NewBufferString("")
		errBuffer = bytes.NewBufferString("")
		logger = boshlog.NewWriterLogger(boshlog.LevelDebug, outBuffer, errBuffer)

		fs = fakeboshsys.NewFakeFileSystem()

		releaseJob = bireljob.Job{
			Name: "fake-job-name",
		}

		renderedJobPath = "fake-path"

		renderedJob = NewRenderedJob(releaseJob, renderedJobPath, fs, logger)
	})

	Describe("Job", func() {
		It("returns the release job", func() {
			Expect(renderedJob.Job()).To(Equal(releaseJob))
		})
	})

	Describe("Path", func() {
		It("returns the rendered job path", func() {
			Expect(renderedJob.Path()).To(Equal(renderedJobPath))
		})
	})

	Describe("Delete", func() {
		It("deletes the rendered job path from the file system", func() {
			err := fs.MkdirAll(renderedJobPath, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			err = renderedJob.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(renderedJobPath)).To(BeFalse())
		})

		Context("when deleting from the file system fails", func() {
			JustBeforeEach(func() {
				fs.RemoveAllError = bosherr.Error("fake-delete-error")
			})

			It("returns an error", func() {
				err := renderedJob.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))
			})
		})
	})

	Describe("DeleteSilently", func() {
		It("deletes the rendered job path from the file system", func() {
			err := fs.MkdirAll(renderedJobPath, os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			renderedJob.DeleteSilently()
			Expect(fs.FileExists(renderedJobPath)).To(BeFalse())
		})

		Context("when deleting from the file system fails", func() {
			JustBeforeEach(func() {
				fs.RemoveAllError = bosherr.Error("fake-delete-error")
			})

			It("logs the error", func() {
				renderedJob.DeleteSilently()

				errorLogString := errBuffer.String()
				Expect(errorLogString).To(ContainSubstring("Failed to delete rendered job"))
				Expect(errorLogString).To(ContainSubstring("fake-delete-error"))
			})
		})
	})
})
