package http

import (
	"net/http"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshretry "github.com/cloudfoundry/bosh-utils/retrystrategy"
	"github.com/cloudfoundry/bosh-utils/errors"
)

type retryClient struct {
	delegate              Client
	maxAttempts           uint
	retryDelay            time.Duration
	logger                boshlog.Logger
	isResponseAttemptable func(*http.Response, error) (bool, error)
}

func NewRetryClient(
	delegate Client,
	maxAttempts uint,
	retryDelay time.Duration,
	logger boshlog.Logger,
) Client {
	return &retryClient{
		delegate:              delegate,
		maxAttempts:           maxAttempts,
		retryDelay:            retryDelay,
		logger:                logger,
		isResponseAttemptable: nil,
	}
}

func NewNetworkSafeRetryClient(
	delegate Client,
	maxAttempts uint,
	retryDelay time.Duration,
	logger boshlog.Logger,
) Client {
	return &retryClient{
		delegate:              delegate,
		maxAttempts:           maxAttempts,
		retryDelay:            retryDelay,
		logger:                logger,
		isResponseAttemptable: func(resp *http.Response,err error) (bool, error) {
			if err != nil || resp.StatusCode == http.StatusGatewayTimeout || resp.StatusCode == http.StatusServiceUnavailable {
				return true, errors.WrapError(err,"Retry")
			}

			directorErrorCodes := []int{400, 401, 403, 404, 500}

			for _, errorCode := range directorErrorCodes {
				if resp.StatusCode == errorCode {
					return false, errors.Errorf("Director responded with non-successful status code HTTP %d. Not attempting to retry...", resp.StatusCode)
				}
			}

			return false, nil
		},
	}
}

func (r *retryClient) Do(req *http.Request) (*http.Response, error) {
	requestRetryable := NewRequestRetryable(req, r.delegate, r.logger, r.isResponseAttemptable)
	retryStrategy := boshretry.NewAttemptRetryStrategy(int(r.maxAttempts), r.retryDelay, requestRetryable, r.logger)
	err := retryStrategy.Try()

	return requestRetryable.Response(), err
}
