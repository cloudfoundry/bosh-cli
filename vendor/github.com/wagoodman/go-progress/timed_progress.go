package progress

import (
	"time"
)

type TimedProgress struct {
	start    time.Time
	duration time.Duration
	complete bool
}

func NewTimedProgress(duration time.Duration) *TimedProgress {
	return &TimedProgress{
		start:    time.Now(),
		duration: duration,
	}
}

func (r *TimedProgress) Current() int64 {
	if r.complete {
		return r.duration.Milliseconds()
	}
	current := time.Since(r.start).Milliseconds()
	if current > r.duration.Milliseconds() {
		r.complete = true
		current = r.duration.Milliseconds()
	}
	return current
}

func (r *TimedProgress) Size() int64 {
	return r.duration.Milliseconds()
}

func (r *TimedProgress) Error() error {
	return nil
}

func (r *TimedProgress) SetCompleted() {
	r.complete = true
}
