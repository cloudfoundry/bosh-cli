package fakes

import (
	"net/http"
)

type attemptOutput struct {
	response    *http.Response
	isRetryable bool
	err         error
}

type FakeRequestRetryable struct {
	attemptOutputs []attemptOutput
	lastResponse   *http.Response

	AttemptCalled int
}

func NewFakeRequestRetryable() *FakeRequestRetryable {
	return &FakeRequestRetryable{
		attemptOutputs: []attemptOutput{},
	}
}

func (r *FakeRequestRetryable) Attempt() (bool, error) {
	r.AttemptCalled++

	currentAttempt := r.attemptOutputs[0]
	r.attemptOutputs = r.attemptOutputs[1:]
	r.lastResponse = currentAttempt.response

	return currentAttempt.isRetryable, currentAttempt.err
}

func (r *FakeRequestRetryable) Response() *http.Response {
	return r.lastResponse
}

func (r *FakeRequestRetryable) AddAttemptBehavior(response *http.Response, isRetryable bool, err error) {
	r.attemptOutputs = append(r.attemptOutputs, attemptOutput{
		response:    response,
		isRetryable: isRetryable,
		err:         err,
	})
}
