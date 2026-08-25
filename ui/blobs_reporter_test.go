package ui_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/ui"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("BlobsReporter", func() {
	var (
		testUI   *testui.Ui
		reporter BlobsReporter
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
		reporter = NewBlobsReporter(testUI)
	})

	Describe("BlobDownloadStarted", func() {
		It("prints download msg", func() {
			reporter.BlobDownloadStarted("path", 100, "blob-id", "blob-sha1")
			Expect(testUI.Said).To(Equal([]string{
				"Blob download 'path' (100 B) (id: blob-id sha1: blob-sha1) started\n"}))
		})
	})

	Describe("BlobDownloadFinished", func() {
		It("prints failed download msg", func() {
			reporter.BlobDownloadFinished("path", "blob-id", errors.New("err"))
			Expect(testUI.Errors).To(Equal([]string{"Blob download 'path' (id: blob-id) failed"}))
		})

		It("prints finished download msg", func() {
			reporter.BlobDownloadFinished("path", "blob-id", nil)
			Expect(testUI.Said).To(Equal([]string{"Blob download 'path' (id: blob-id) finished\n"}))
		})
	})

	Describe("BlobUploadStarted", func() {
		It("prints upload msg", func() {
			reporter.BlobUploadStarted("path", 100, "blob-sha1")
			Expect(testUI.Said).To(Equal([]string{"Blob upload 'path' (100 B) (sha1: blob-sha1) started\n"}))
		})
	})

	Describe("BlobUploadFinished", func() {
		It("prints failed upload msg", func() {
			reporter.BlobUploadFinished("path", "", errors.New("err"))
			Expect(testUI.Errors).To(Equal([]string{"Blob upload 'path' failed"}))
		})

		It("prints finished upload msg", func() {
			reporter.BlobUploadFinished("path", "blob-id", nil)
			Expect(testUI.Said).To(Equal([]string{"Blob upload 'path' (id: blob-id) finished\n"}))
		})
	})
})
