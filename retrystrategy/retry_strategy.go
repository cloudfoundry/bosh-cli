package retrystrategy

type RetryStrategy interface {
	Try() error
}

type Retryable interface {
	Attempt() error
}

type retryable struct {
	attemptFunc func() error
}

func (r *retryable) Attempt() error {
	return r.attemptFunc()
}

func NewRetryable(attemptFunc func() error) Retryable {
	return &retryable{
		attemptFunc: attemptFunc,
	}
}
