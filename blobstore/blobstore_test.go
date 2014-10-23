package blobstore_test

import (
	"errors"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeboshdavcli "github.com/cloudfoundry/bosh-agent/davcli/client/fakes"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/blobstore"
)

var _ = Describe("Blobstore", func() {
	var (
		blobstore     Blobstore
		fakeDavClient *fakeboshdavcli.FakeClient
		fs            *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeDavClient = fakeboshdavcli.NewFakeClient()
		fs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		blobstore = NewBlobstore(fakeDavClient, fs, logger)
	})

	Describe("Get", func() {
		It("gets the blob from the blobstore", func() {
			fakeDavClient.GetContents = ioutil.NopCloser(strings.NewReader("fake-content"))

			err := blobstore.Get("fake-blob-id", "fake-destination-path")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeDavClient.GetPath).To(Equal("fake-blob-id"))
		})

		It("saves the blob to the destination path", func() {
			fakeDavClient.GetContents = ioutil.NopCloser(strings.NewReader("fake-content"))

			err := blobstore.Get("fake-blob-id", "fake-destination-path")
			Expect(err).ToNot(HaveOccurred())

			contents, err := fs.ReadFileString("fake-destination-path")
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal("fake-content"))
		})

		Context("when getting from blobstore fails", func() {
			It("returns an error", func() {
				fakeDavClient.GetErr = errors.New("fake-get-error")
				err := blobstore.Get("fake-blob-id", "fake-destination-path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-get-error"))
			})
		})
	})

	Describe("Save", func() {
		BeforeEach(func() {
			fs.RegisterOpenFile("fake-source-path", &fakesys.FakeFile{
				Contents: []byte("fake-contents"),
			})
		})

		It("saves blob to blobstore", func() {
			err := blobstore.Save("fake-source-path", "fake-blob-id")
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeDavClient.PutPath).To(Equal("fake-blob-id"))
			Expect(fakeDavClient.PutContents).To(Equal("fake-contents"))
		})
	})
})
