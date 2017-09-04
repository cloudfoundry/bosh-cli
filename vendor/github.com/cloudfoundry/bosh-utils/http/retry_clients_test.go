package http_test

import (
	"net/http"

	fakehttp "github.com/cloudfoundry/bosh-utils/http/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"

	. "github.com/cloudfoundry/bosh-utils/http"
)

var _ = Describe("RetryClients", func() {

	Describe("RetryClient", func() {
		Describe("Do", func() {
			var (
				retryClient Client
				maxAttempts int
				fakeClient  *fakehttp.FakeClient
			)

			BeforeEach(func() {
				fakeClient = fakehttp.NewFakeClient()
				logger := boshlog.NewLogger(boshlog.LevelNone)
				maxAttempts = 7

				retryClient = NewRetryClient(fakeClient, uint(maxAttempts), 0, logger)
			})

			It("returns response from retryable request", func() {
				fakeClient.SetMessage("fake-response-body")
				fakeClient.StatusCode = 204

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				Expect(readString(resp.Body)).To(Equal("fake-response-body"))
				Expect(resp.StatusCode).To(Equal(204))
			})

			It("attemps once if request is successful", func() {
				fakeClient.StatusCode = 200

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				Expect(fakeClient.CallCount).To(Equal(1))
				Expect(fakeClient.Requests).To(ContainElement(req))
			})

			It("retries for maxAttempts if request is failing", func() {
				fakeClient.StatusCode = 404

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).To(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(404))

				Expect(fakeClient.CallCount).To(Equal(maxAttempts))
				Expect(fakeClient.Requests).To(ContainElement(req))
			})
		})
	})

	Describe("NetworkSafeClient", func() {
		Describe("Do", func() {
			var (
				retryClient Client
				maxAttempts int
				fakeClient  *fakehttp.FakeClient
			)

			BeforeEach(func() {
				fakeClient = fakehttp.NewFakeClient()
				logger := boshlog.NewLogger(boshlog.LevelNone)
				maxAttempts = 7

				retryClient = NewNetworkSafeRetryClient(fakeClient, uint(maxAttempts), 0, logger)
			})

			It("returns response from retryable request", func() {
				fakeClient.SetMessage("fake-response-body")
				fakeClient.StatusCode = 204

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())

				Expect(readString(resp.Body)).To(Equal("fake-response-body"))
				Expect(resp.StatusCode).To(Equal(204))
			})

			directorErrorCodes := []int{400, 401, 403, 404, 500}
			for _, code := range directorErrorCodes {
				It(fmt.Sprintf("attemps once if request is %d", code), func() {
					fakeClient.StatusCode = code

					req := &http.Request{}
					resp, err := retryClient.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(code))

					Expect(fakeClient.CallCount).To(Equal(1))
					Expect(fakeClient.Requests).To(ContainElement(req))
				})
			}

			for code := 200; code < 400; code++ {
				successHttpCode := code
				It(fmt.Sprintf("attempts once if request is %d", code), func() {
					fakeClient.StatusCode = successHttpCode

					req := &http.Request{}
					resp, err := retryClient.Do(req)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(successHttpCode))

					Expect(fakeClient.CallCount).To(Equal(1))
					Expect(fakeClient.Requests).To(ContainElement(req))
				})
			}

			Context("underlying connection errors should not be influenced by request method", func() {
				for _, method := range []string{"GET", "HEAD", "POST", "DELETE"} {
					It(fmt.Sprintf("retries for maxAttempts with a %s request", method), func() {
						fakeClient.SetNilResponse()
						fakeClient.Error = errors.New("fake-err")

						req := &http.Request{Method: method}
						resp, err := retryClient.Do(req)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
						Expect(resp).To(BeNil())

						Expect(fakeClient.CallCount).To(Equal(maxAttempts))
						Expect(fakeClient.Requests).To(ContainElement(req))
					})
				}
			})

			timeoutCodes := []int{
				http.StatusGatewayTimeout,
				http.StatusServiceUnavailable,
			}
			for _, code := range timeoutCodes {
				for _, method := range []string{"GET", "HEAD"} {
					Context(fmt.Sprintf("timeout http status code '%d' with %s request", code, method), func() {
						It("retries for maxAttempts", func() {
							fakeClient.StatusCode = code

							req := &http.Request{Method: method}
							resp, err := retryClient.Do(req)
							Expect(err).To(HaveOccurred())

							Expect(resp.StatusCode).To(Equal(code))

							Expect(fakeClient.CallCount).To(Equal(maxAttempts))
							Expect(fakeClient.Requests).To(ContainElement(req))
						})
					})
				}

				for _, method := range []string{"POST", "DELETE"} {
					Context(fmt.Sprintf("timeout http status code '%d' with %s request", code, method), func() {
						It("does not retry", func() {
							fakeClient.StatusCode = code

							req := &http.Request{Method: method}
							resp, err := retryClient.Do(req)
							Expect(err).ToNot(HaveOccurred())

							Expect(resp.StatusCode).To(Equal(code))

							Expect(fakeClient.CallCount).To(Equal(1))
							Expect(fakeClient.Requests).To(ContainElement(req))
						})
					})
				}
			}

		})

	})

})
