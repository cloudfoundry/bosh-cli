package retrystrategy

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
)

type timeoutRetryStrategy struct {
	timeout     time.Duration
	delay       time.Duration
	retryable   Retryable
	timeService boshtime.Service
	logger      boshlog.Logger
	logTag      string
}

func NewTimeoutRetryStrategy(
	timeout time.Duration,
	delay time.Duration,
	retryable Retryable,
	timeService boshtime.Service,
	logger boshlog.Logger,
) RetryStrategy {
	return &timeoutRetryStrategy{
		timeout:     timeout,
		delay:       delay,
		retryable:   retryable,
		timeService: timeService,
		logger:      logger,
		logTag:      "timeoutRetryStrategy",
	}
}

func (s *timeoutRetryStrategy) Try() error {
	var err error
	var isRetryable bool

	now := s.timeService.Now()
	deadline := now.Add(s.timeout)
	deadlineMinusDelay := deadline.Add(-1 * s.delay)

	for i := 0; true; i++ {
		s.logger.Debug(s.logTag, "Making attempt #%d", i)

		isRetryable, err = s.retryable.Attempt()
		if err == nil {
			return nil
		}

		if !isRetryable {
			return err
		}

		now = s.timeService.Now()
		if now.After(deadline) || now.After(deadlineMinusDelay) {
			return err
		}

		s.timeService.Sleep(s.delay)
	}

	return err
}
