# workpool

Import this package via `code.cloudfoundry.org/workpool`.

Use a `WorkPool` to perform units of work concurrently at a maximum rate. The worker goroutines will increase to the maximum number of workers as work requires it, and gradually decrease to 0 if unused.

A `Throttler` performs a specified batch of work at a given maximum rate, internally creating a `WorkPool` and then stopping it when done.

## Reporting issues and requesting features

Please report all issues and feature requests in [cloudfoundry/diego-release](https://github.com/cloudfoundry/diego-release/issues).

## Example

```go
type RateLimitingHandler struct {
	backend http.Handler
	pool    *workpool.WorkPool
}

// Ensures that the backend handler never processes more than maxInFlight requests at a time
func NewRateLimitingHandler(backend http.Handler, maxInFlight int) (*RateLimitingHandler, error) {
	pool, err := workpool.NewWorkPool(maxInFlight)
	if err != nil {
		return nil, err
	}

	return &RateLimitingHandler{
		backend: backend,
		pool:    pool,
	}, nil
}

func (rh *RateLimitingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rh.pool.Submit(func() {
		rh.backend.ServeHTTP(w, req)
	})
}
```
