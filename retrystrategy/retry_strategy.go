package retrystrategy

type RetryStrategy interface {
	Try() error
}

type Retryable interface {
	Attempt() error
}
