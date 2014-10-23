package http

import (
	"net/http"
)

type RetryStrategy interface {
	NewRetryHandler() RetryHandler
}

type RetryHandler interface {
	IsRetryable(req *http.Request, resp *http.Response, err error) bool
}
