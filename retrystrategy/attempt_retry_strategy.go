package retrystrategy

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

type attemptRetryStrategy struct {
	maxAttempts int
	delay       time.Duration
	retryable   Retryable
	logger      boshlog.Logger
	logTag      string
}

func NewAttemptRetryStrategy(
	maxAttempts int,
	delay time.Duration,
	retryable Retryable,
	logger boshlog.Logger,
) RetryStrategy {
	return &attemptRetryStrategy{
		maxAttempts: maxAttempts,
		delay:       delay,
		retryable:   retryable,
		logger:      logger,
		logTag:      "attemptRetryStrategy",
	}
}

func (s *attemptRetryStrategy) Try() error {
	var err error
	for i := 0; i < s.maxAttempts; i++ {
		s.logger.Debug(s.logTag, "Making attempt #%d", i)
		err = s.retryable.Attempt()
		if err == nil {
			return nil
		}
		time.Sleep(s.delay)
	}

	return err
}
