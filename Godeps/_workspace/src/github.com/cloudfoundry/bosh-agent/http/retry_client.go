package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshtime "github.com/cloudfoundry/bosh-agent/time"
	boshuuid "github.com/cloudfoundry/bosh-agent/uuid"
)

type retryClient struct {
	delegate      Client
	retryStrategy RetryStrategy
	delay         time.Duration
	timeService   boshtime.Service
	logger        boshlog.Logger
	logTag        string
	uuidGenerator boshuuid.Generator
}

func NewRetryClient(
	delegate Client,
	retryStrategy RetryStrategy,
	delay time.Duration,
	timeService boshtime.Service,
	logger boshlog.Logger,
) Client {
	return &retryClient{
		delegate:      delegate,
		retryStrategy: retryStrategy,
		delay:         delay,
		timeService:   timeService,
		logger:        logger,
		logTag:        "retry-client",
		uuidGenerator: boshuuid.NewGenerator(),
	}
}

func (r *retryClient) Do(req *http.Request) (*http.Response, error) {
	attempt := 1
	requestID, err := r.uuidGenerator.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating request uuid")
	}

	// buffer request body to memory for retries
	var bodyBytes []byte

	if req.Body != nil {
		bodyBytes, err = ReadAndClose(req.Body)
		if err != nil {
			return nil, bosherr.WrapError(err, "Buffering request body")
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}

	r.logger.Debug(r.logTag, "[requestID=%s] Requesting (attempt=%d): %s", requestID, attempt, r.formatRequest(req))
	resp, err := r.delegate.Do(req)
	if r.wasSuccessful(resp, err) {
		r.logger.Debug(r.logTag, "[requestID=%s] Request succeeded (attempt=%d): %s", requestID, attempt, r.formatResponse(resp))
		return resp, err
	}

	retryHandler := r.retryStrategy.NewRetryHandler()

	for {
		if !retryHandler.IsRetryable(req, resp, err) {
			r.logger.Warn(r.logTag, "[requestID=%s] Request failed (attempts=%d): %s", requestID, attempt, r.formatResponse(resp))
			return resp, err
		}

		r.logger.Debug(r.logTag, "[requestID=%s] Request attempt failed (attempts=%d, sleep=%s): %s", requestID, attempt, r.delay, r.formatResponse(resp))

		attempt = attempt + 1
		r.timeService.Sleep(r.delay)

		// reset request body, because readers cannot be re-read
		if bodyBytes != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
		}

		r.logger.Debug(r.logTag, "[requestID=%s] Requesting (attempt=%d)", requestID, attempt)
		resp, err = r.delegate.Do(req)
		if r.wasSuccessful(resp, err) {
			r.logger.Debug(r.logTag, "[requestID=%s] Request succeeded (attempts=%d): %s", requestID, attempt, r.formatResponse(resp))
			return resp, err
		}
	}
}

func (r *retryClient) wasSuccessful(resp *http.Response, err error) bool {
	return err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (r *retryClient) formatRequest(req *http.Request) string {
	if req == nil {
		return "Request(nil)"
	}

	return fmt.Sprintf("Request{ Method: '%s', URL: '%s' }", req.Method, req.URL)
}

func (r *retryClient) formatResponse(resp *http.Response) string {
	if resp == nil {
		return "Response(nil)"
	}

	return fmt.Sprintf("Response{ StatusCode: %d, Status: '%s' }", resp.StatusCode, resp.Status)
}
