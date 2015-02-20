package retrystrategy_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"

	. "github.com/cloudfoundry/bosh-agent/retrystrategy"
)

var _ = Describe("TimeoutRetryStrategy", func() {
	var (
		fakeTimeService *faketime.FakeService
		logger          boshlog.Logger
	)

	BeforeEach(func() {
		fakeTimeService = &faketime.FakeService{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
	})

	Describe("Try", func() {
		Context("when there are errors during a try", func() {
			It("retries until the timeout", func() {
				now := time.Now()
				fakeTimeService.NowTimes = []time.Time{
					now,
					now.Add(10 * time.Second), // after 1st attempt
					now.Add(30 * time.Second), // after 2nd attempt
					now.Add(50 * time.Second), // after 3rd attempt
				}

				retryable := newSimpleRetryable([]attemptOutput{
					{
						IsRetryable: true,
						AttemptErr:  errors.New("first-error"),
					},
					{
						IsRetryable: true,
						AttemptErr:  errors.New("second-error"),
					},
					{
						IsRetryable: true,
						AttemptErr:  errors.New("third-error"),
					},
					{
						IsRetryable: true,
						AttemptErr:  errors.New("fourth-error"),
					},
				})
				// deadline between 2nd and 3rd attempts
				timeoutRetryStrategy := NewTimeoutRetryStrategy(40*time.Second, 5*time.Second, retryable, fakeTimeService, logger)
				err := timeoutRetryStrategy.Try()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("third-error"))
				Expect(retryable.Attempts).To(Equal(3))
			})

			It("stops without a trailing delay", func() {
				now := time.Now()
				fakeTimeService.NowTimes = []time.Time{
					now,
					now.Add(10 * time.Second), // after 1st attempt
					now.Add(30 * time.Second), // after 2nd attempt
					now.Add(50 * time.Second), // after 3rd attempt
				}

				retryable := newSimpleRetryable([]attemptOutput{
					{
						IsRetryable: true,
						AttemptErr:  errors.New("first-error"),
					},
					{
						IsRetryable: true,
						AttemptErr:  errors.New("second-error"),
					},
					{
						IsRetryable: true,
						AttemptErr:  errors.New("third-error"),
					},
				})
				// deadline after 2nd attempt errors, but (deadline - delay) between 2nd and 3rd attempts
				timeoutRetryStrategy := NewTimeoutRetryStrategy(35*time.Second, 10*time.Second, retryable, fakeTimeService, logger)
				err := timeoutRetryStrategy.Try()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("second-error"))
				Expect(retryable.Attempts).To(Equal(2))
			})
		})

		Context("when the attempt stops being retryable", func() {
			It("stops trying", func() {
				now := time.Now()
				fakeTimeService.NowTimes = []time.Time{
					now,
					now.Add(1 * time.Second),
				}

				retryable := newSimpleRetryable([]attemptOutput{
					{
						IsRetryable: true,
						AttemptErr:  errors.New("first-error"),
					},
					{
						IsRetryable: false,
						AttemptErr:  errors.New("second-error"),
					},
				})
				timeoutRetryStrategy := NewTimeoutRetryStrategy(3*time.Second, 1*time.Second, retryable, fakeTimeService, logger)
				err := timeoutRetryStrategy.Try()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("second-error"))
				Expect(retryable.Attempts).To(Equal(2))
			})
		})

		Context("when there are no errors", func() {
			It("does not retry", func() {
				now := time.Now()
				fakeTimeService.NowTimes = []time.Time{
					now,
				}

				retryable := newSimpleRetryable([]attemptOutput{
					{
						IsRetryable: true,
						AttemptErr:  nil,
					},
				})
				timeoutRetryStrategy := NewTimeoutRetryStrategy(3*time.Second, 1*time.Second, retryable, fakeTimeService, logger)
				err := timeoutRetryStrategy.Try()
				Expect(err).ToNot(HaveOccurred())
				Expect(retryable.Attempts).To(Equal(1))
			})
		})
	})
})
