package http_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "github.com/cloudfoundry/bosh-agent/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakehttp "github.com/cloudfoundry/bosh-agent/http/fakes"
	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

var _ = Describe("RetryClient", func() {

	Describe("Do", func() {

		var (
			retryClient  Client
			fakeClient   *fakehttp.FakeClient
			fakeStrategy *fakehttp.FakeRetryStrategy
			delay        time.Duration
			timeService  *faketime.FakeService
		)

		BeforeEach(func() {
			fakeClient = fakehttp.NewFakeClient()
			fakeStrategy = fakehttp.NewFakeRetryStrategy()
			delay = 1 * time.Millisecond
			timeService = &faketime.FakeService{}
			logger := boshlog.NewLogger(boshlog.LevelNone)

			retryClient = NewRetryClient(fakeClient, fakeStrategy, delay, timeService, logger)
		})

		It("calls Do on the delegate", func() {
			fakeStrategy.AddIsRetryableReturn(false)
			fakeClient.SetMessage("fake-response-body")
			fakeClient.StatusCode = 200

			req := &http.Request{}
			resp, err := retryClient.Do(req)
			Expect(err).ToNot(HaveOccurred())
			Expect(readString(resp.Body)).To(Equal("fake-response-body"))
			Expect(resp.StatusCode).To(Equal(200))

			Expect(fakeClient.CallCount).To(Equal(1))
			Expect(fakeClient.Requests).To(ContainElement(req))
		})

		Context("when request returns an error", func() {
			BeforeEach(func() {
				fakeClient.Error = errors.New("fake-response-error")
			})

			It("retries errors until the retry strategy says to stop", func() {
				for i := 0; i < 5; i++ {
					fakeStrategy.AddIsRetryableReturn(true)
				}
				fakeStrategy.AddIsRetryableReturn(false)

				req := &http.Request{}
				_, err := retryClient.Do(req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-response-error"))

				Expect(fakeClient.CallCount).To(Equal(6))
			})

			It("sleeps before retrying", func() {
				fakeStrategy.AddIsRetryableReturn(true)
				fakeStrategy.AddIsRetryableReturn(false)

				req := &http.Request{}
				_, err := retryClient.Do(req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-response-error"))

				Expect(fakeClient.CallCount).To(Equal(2))
				Expect(len(timeService.SleepInputs)).To(Equal(1))
				Expect(timeService.SleepInputs[0]).To(Equal(delay))
			})
		})

		Context("when response status code is not between 200 and 300", func() {
			BeforeEach(func() {
				fakeClient.SetMessage("fake-response-body")
				fakeClient.StatusCode = 404
			})

			It("retries errors until the retry strategy says to stop", func() {
				for i := 0; i < 5; i++ {
					fakeStrategy.AddIsRetryableReturn(true)
				}
				fakeStrategy.AddIsRetryableReturn(false)

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(readString(resp.Body)).To(Equal("fake-response-body"))
				Expect(resp.StatusCode).To(Equal(404))

				Expect(fakeClient.CallCount).To(Equal(6))
			})

			It("sleeps before retrying", func() {
				fakeStrategy.AddIsRetryableReturn(true)
				fakeStrategy.AddIsRetryableReturn(false)

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(readString(resp.Body)).To(Equal("fake-response-body"))
				Expect(resp.StatusCode).To(Equal(404))

				Expect(fakeClient.CallCount).To(Equal(2))
				Expect(len(timeService.SleepInputs)).To(Equal(1))
				Expect(timeService.SleepInputs[0]).To(Equal(delay))
			})

			It("re-populates the request body on subsequent attempts", func() {
				fakeStrategy.AddIsRetryableReturn(true)
				fakeStrategy.AddIsRetryableReturn(false)

				req := &http.Request{
					Body: ioutil.NopCloser(strings.NewReader("fake-request-body")),
				}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(readString(resp.Body)).To(Equal("fake-response-body"))
				Expect(resp.StatusCode).To(Equal(404))

				Expect(fakeClient.RequestBodies[0]).To(Equal("fake-request-body"))
				Expect(fakeClient.RequestBodies[1]).To(Equal("fake-request-body"))
			})
		})

		Context("when the request succeeds after failing", func() {
			BeforeEach(func() {
				errorResponse := &http.Response{
					StatusCode: 500,
					Body:       NewStringReadCloser("fake-error-body"),
				}
				fakeClient.AddDoBehavior(errorResponse, nil)

				successResponse := &http.Response{
					StatusCode: 200,
					Body:       NewStringReadCloser("fake-success-body"),
				}
				fakeClient.AddDoBehavior(successResponse, nil)
			})

			It("retries errors until the retry strategy says to stop", func() {
				fakeStrategy.AddIsRetryableReturn(true)

				req := &http.Request{}
				resp, err := retryClient.Do(req)
				Expect(err).ToNot(HaveOccurred())
				Expect(readString(resp.Body)).To(Equal("fake-success-body"))
				Expect(resp.StatusCode).To(Equal(200))

				Expect(fakeClient.CallCount).To(Equal(2))
			})
		})
	})
})

func readString(body io.ReadCloser) string {
	content, err := ReadAndClose(body)
	Expect(err).ToNot(HaveOccurred())
	return string(content)
}
