package http

import (
	"net/http"
)

type AttemptRetryStrategy struct {
	maxAttempts uint
}

func NewAttemptRetryStrategy(maxAttempts uint) RetryStrategy {
	return &AttemptRetryStrategy{
		maxAttempts: maxAttempts,
	}
}

func (a *AttemptRetryStrategy) NewRetryHandler() RetryHandler {
	return NewAttemptRetryHandler(a.maxAttempts)
}

type attemptRetryHandler struct {
	attempts    uint
	maxAttempts uint
}

func NewAttemptRetryHandler(maxAttempts uint) RetryHandler {
	return &attemptRetryHandler{maxAttempts: maxAttempts}
}

func (a *attemptRetryHandler) IsRetryable(_ *http.Request, _ *http.Response, _ error) bool {
	a.attempts = a.attempts + 1
	return a.attempts <= a.maxAttempts
}
