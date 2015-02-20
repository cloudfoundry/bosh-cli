package time

import (
	"time"
)

type concreteService struct{}

func NewConcreteService() Service {
	return concreteService{}
}

func (s concreteService) Now() time.Time {
	return time.Now()
}

func (s concreteService) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
