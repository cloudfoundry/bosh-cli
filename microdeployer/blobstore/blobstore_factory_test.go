package blobstore_test

import (
	"crypto/tls"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshdavcli "github.com/cloudfoundry/bosh-agent/davcli/client"
	boshdavcliconf "github.com/cloudfoundry/bosh-agent/davcli/config"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/microdeployer/blobstore"
)

var _ = Describe("BlobstoreFactory", func() {
	var (
		httpClient       http.Client
		fs               *fakesys.FakeFileSystem
		logger           boshlog.Logger
		blobstoreFactory Factory
	)
	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient = http.Client{Transport: tr}

		blobstoreFactory = NewBlobstoreFactory(fs, logger)
	})

	Describe("Create", func() {
		Context("when username and password are provided", func() {
			It("returns the blobstore", func() {
				blobstore, err := blobstoreFactory.Create("https://fake-user:fake-password@fake-host:1234")
				Expect(err).ToNot(HaveOccurred())
				davClient := boshdavcli.NewClient(boshdavcliconf.Config{
					Endpoint: "https://fake-host:1234/blobs",
					User:     "fake-user",
					Password: "fake-password",
				}, &httpClient)
				expectedBlobstore := NewBlobstore(davClient, fs, logger)
				Expect(blobstore).To(Equal(expectedBlobstore))
			})
		})

		Context("when URL does not have username and password", func() {
			// This test was added because parsing password is failing when userInfo is missing in URL
			It("returns the blobstore", func() {
				davClient := boshdavcli.NewClient(boshdavcliconf.Config{
					Endpoint: "https://fake-host:1234/blobs",
					User:     "",
					Password: "",
				}, &httpClient)
				expectedBlobstore := NewBlobstore(davClient, fs, logger)

				blobstore, err := blobstoreFactory.Create("https://fake-host:1234")
				Expect(err).ToNot(HaveOccurred())
				Expect(blobstore).To(Equal(expectedBlobstore))
			})
		})
	})
})
