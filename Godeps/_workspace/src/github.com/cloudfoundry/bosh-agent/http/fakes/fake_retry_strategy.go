package fakes

import (
	"net/http"

	boshhttp "github.com/cloudfoundry/bosh-agent/http"
)

type FakeRetryStrategy struct {
	isRetryableReturns []bool
}

func NewFakeRetryStrategy() *FakeRetryStrategy {
	return &FakeRetryStrategy{}
}

func (rs *FakeRetryStrategy) NewRetryHandler() boshhttp.RetryHandler {
	return &retryHandler{retryStrategy: rs}
}

func (rs *FakeRetryStrategy) AddIsRetryableReturn(isRetryable bool) {
	rs.isRetryableReturns = append(rs.isRetryableReturns, isRetryable)
}

type retryHandler struct {
	retryStrategy *FakeRetryStrategy
}

func (rh retryHandler) IsRetryable(*http.Request, *http.Response, error) bool {
	isRetryable := rh.retryStrategy.isRetryableReturns[0]
	rh.retryStrategy.isRetryableReturns = rh.retryStrategy.isRetryableReturns[1:]
	return isRetryable
}
