package retrystrategy_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
)

var _ = Describe("AttemptRetryStrategy", func() {
	var (
		logger boshlog.Logger
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
	})

	Describe("Try", func() {
		Context("when there are errors during a try", func() {
			It("retries until the max attempts are used up", func() {
				retryable := NewSimpleRetryable([]error{
					errors.New("first-error"),
					errors.New("second-error"),
					errors.New("third-error"),
				})
				attemptRetryStrategy := NewAttemptRetryStrategy(3, retryable, logger)
				err := attemptRetryStrategy.Try()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("third-error"))
				Expect(retryable.Attempts).To(Equal(3))
			})
		})

		Context("when there are no errors", func() {
			It("does not retry", func() {
				retryable := NewSimpleRetryable([]error{})
				attemptRetryStrategy := NewAttemptRetryStrategy(3, retryable, logger)
				err := attemptRetryStrategy.Try()
				Expect(err).ToNot(HaveOccurred())
				Expect(retryable.Attempts).To(Equal(1))
			})
		})
	})
})

type simpleRetryable struct {
	attemptErrors []error
	Attempts      int
}

func NewSimpleRetryable(attemptErrors []error) *simpleRetryable {
	return &simpleRetryable{
		attemptErrors: attemptErrors,
	}
}

func (r *simpleRetryable) Attempt() error {
	r.Attempts++

	if len(r.attemptErrors) > 0 {
		attemptError := r.attemptErrors[0]
		r.attemptErrors = r.attemptErrors[1:]
		return attemptError
	}

	return nil
}
