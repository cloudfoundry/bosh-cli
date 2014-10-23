package http_test

import (
	"errors"
	"net/http"

	. "github.com/cloudfoundry/bosh-agent/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AttemptRetryStrategy", func() {
	var (
		attemptRetryStrategy RetryStrategy
		attemptRetryHandler  RetryHandler
		maxAttempts          uint
	)

	BeforeEach(func() {
		maxAttempts = 2
		attemptRetryStrategy = NewAttemptRetryStrategy(maxAttempts)
		attemptRetryHandler = attemptRetryStrategy.NewRetryHandler()
	})

	Describe("IsRetryable", func() {
		var (
			request  *http.Request
			response *http.Response
			err      error
		)

		BeforeEach(func() {
			request = &http.Request{}
			response = &http.Response{}
			err = nil
		})

		It("retries completed requests until maxAttempts are exhausted", func() {
			for i := uint(0); i < maxAttempts; i++ {
				Expect(attemptRetryHandler.IsRetryable(request, response, err)).To(BeTrue())
			}

			Expect(attemptRetryHandler.IsRetryable(request, response, err)).To(BeFalse())
		})

		It("retries errored requests until maxAttempts are exhausted", func() {
			err = errors.New("fake-error")

			for i := uint(0); i < maxAttempts; i++ {
				Expect(attemptRetryHandler.IsRetryable(request, response, err)).To(BeTrue())
			}

			Expect(attemptRetryHandler.IsRetryable(request, response, err)).To(BeFalse())
		})

	})
})
