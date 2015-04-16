package sshtunnel

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"time"

	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"
)

var _ = Describe("SSH", func() {
	Describe("SSHRetryStrategy", func() {
		var (
			sshRetryStrategy         *SSHRetryStrategy
			fakeTimeService          *faketime.FakeService
			connectionRefusedTimeout time.Duration
			authFailureTimeout       time.Duration
			startTime                time.Time
		)

		BeforeEach(func() {
			startTime = time.Now()
			fakeTimeService = &faketime.FakeService{
				NowTimes: []time.Time{startTime},
			}
			connectionRefusedTimeout = 10 * time.Minute
			authFailureTimeout = 5 * time.Minute

			sshRetryStrategy = &SSHRetryStrategy{
				ConnectionRefusedTimeout: connectionRefusedTimeout,
				AuthFailureTimeout:       authFailureTimeout,
				TimeService:              fakeTimeService,
			}
		})

		Describe("IsRetryable", func() {
			refusedErr := errors.New("connection refused")
			authErr := errors.New("unable to authenticate")

			Context("when err is connection refused", func() {
				It("retries for connectionRefusedTimeout", func() {
					fakeTimeService.NowTimes = []time.Time{startTime}
					Expect(sshRetryStrategy.IsRetryable(refusedErr)).To(BeTrue())

					fakeTimeService.NowTimes = []time.Time{startTime.Add(connectionRefusedTimeout).Add(-1 * time.Second)}
					Expect(sshRetryStrategy.IsRetryable(refusedErr)).To(BeTrue())

					fakeTimeService.NowTimes = []time.Time{startTime.Add(connectionRefusedTimeout)}
					Expect(sshRetryStrategy.IsRetryable(refusedErr)).To(BeFalse())
				})
			})

			Context("when err is unable to authenticate", func() {
				It("retries for authFailureTimeout", func() {
					fakeTimeService.NowTimes = []time.Time{startTime}
					Expect(sshRetryStrategy.IsRetryable(authErr)).To(BeTrue())

					fakeTimeService.NowTimes = []time.Time{startTime.Add(authFailureTimeout).Add(-1 * time.Second)}
					Expect(sshRetryStrategy.IsRetryable(authErr)).To(BeTrue())

					fakeTimeService.NowTimes = []time.Time{startTime.Add(authFailureTimeout)}
					Expect(sshRetryStrategy.IsRetryable(authErr)).To(BeFalse())
				})
			})

			Context("when connection is refused, then err becomes unable to authenticate", func() {
				It("retries for connectionRefusedTimeout", func() {
					fakeTimeService.NowTimes = []time.Time{startTime}
					Expect(sshRetryStrategy.IsRetryable(refusedErr)).To(BeTrue())

					fakeTimeService.NowTimes = []time.Time{startTime.Add(1 * time.Minute)}
					Expect(sshRetryStrategy.IsRetryable(refusedErr)).To(BeTrue())

					lastConnectionRefusedTime := startTime.Add(1 * time.Minute)
					fakeTimeService.NowTimes = []time.Time{lastConnectionRefusedTime.Add(authFailureTimeout).Add(-1 * time.Second)}
					Expect(sshRetryStrategy.IsRetryable(authErr)).To(BeTrue())

					fakeTimeService.NowTimes = []time.Time{lastConnectionRefusedTime.Add(authFailureTimeout)}
					Expect(sshRetryStrategy.IsRetryable(authErr)).To(BeFalse())
				})
			})

			It("'no common algorithms' error fails immediately", func() {
				fakeTimeService.NowTimes = []time.Time{startTime}
				Expect(sshRetryStrategy.IsRetryable(errors.New("no common algorithms"))).To(BeFalse())
			})

			It("all other errors fail after the connection refused timeout", func() {
				fakeTimeService.NowTimes = []time.Time{startTime}
				Expect(sshRetryStrategy.IsRetryable(errors.New("another error"))).To(BeTrue())

				fakeTimeService.NowTimes = []time.Time{startTime.Add(connectionRefusedTimeout)}
				Expect(sshRetryStrategy.IsRetryable(errors.New("another error"))).To(BeFalse())
			})
		})
	})
})
