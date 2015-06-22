package client_test

import (
	"errors"
	"io/ioutil"
	"strings"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-davcli/client"
	davconf "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-davcli/config"
	fakehttp "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/http/fakes"
)

var _ = Describe("Client", func() {
	var (
		fakeHTTPClient *fakehttp.FakeClient
		config         davconf.Config
		client         Client
	)

	BeforeEach(func() {
		fakeHTTPClient = fakehttp.NewFakeClient()
		client = NewClient(config, fakeHTTPClient)
	})

	Describe("Get", func() {
		It("returns the response body from the given path", func() {
			fakeHTTPClient.StatusCode = 200
			fakeHTTPClient.SetMessage("response")

			responseBody, err := client.Get("/")
			Expect(err).NotTo(HaveOccurred())
			buf := make([]byte, 1024)
			n, _ := responseBody.Read(buf)
			Expect(string(buf[0:n])).To(Equal("response"))
		})

		Context("when the http request fails", func() {
			BeforeEach(func() {
				fakeHTTPClient.Error = errors.New("")
			})

			It("returns err", func() {
				responseBody, err := client.Get("/")
				Expect(responseBody).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting dav blob /"))
			})
		})

		Context("when the http response code is not 200", func() {
			BeforeEach(func() {
				fakeHTTPClient.StatusCode = 300
				fakeHTTPClient.SetMessage("response")
			})

			It("returns err", func() {
				responseBody, err := client.Get("/")
				Expect(responseBody).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting dav blob /: Wrong response code: 300; body: response"))
			})
		})
	})

	Describe("Put", func() {
		Context("When the put request succeeds", func() {
			itUploadsABlob := func() {
				body := ioutil.NopCloser(strings.NewReader("content"))
				err := client.Put("/", body, int64(7))
				Expect(err).NotTo(HaveOccurred())
				Expect(len(fakeHTTPClient.Requests)).To(Equal(1))
				req := fakeHTTPClient.Requests[0]
				Expect(req.ContentLength).To(Equal(int64(7)))
				Expect(fakeHTTPClient.RequestBodies).To(Equal([]string{"content"}))
			}

			It("uploads the given content if the blob does not exist", func() {
				fakeHTTPClient.StatusCode = 201
				itUploadsABlob()
			})

			It("uploads the given content if the blob exists", func() {
				fakeHTTPClient.StatusCode = 204
				itUploadsABlob()
			})
		})

		Context("when the http request fails", func() {
			BeforeEach(func() {
				fakeHTTPClient.Error = errors.New("")
			})

			It("returns err", func() {
				body := ioutil.NopCloser(strings.NewReader("content"))
				err := client.Put("/", body, int64(7))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Putting dav blob /"))
			})
		})

		Context("when the http response code is not 201 or 204", func() {
			BeforeEach(func() {
				fakeHTTPClient.StatusCode = 300
				fakeHTTPClient.SetMessage("response")
			})

			It("returns err", func() {
				body := ioutil.NopCloser(strings.NewReader("content"))
				err := client.Put("/", body, int64(7))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Putting dav blob /: Wrong response code: 300; body: response"))
			})
		})
	})
})
