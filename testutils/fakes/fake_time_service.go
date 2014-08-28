package fakes

import (
	"time"
)

type FakeTimeService struct {
	NowTimes []time.Time
}

func (f *FakeTimeService) Now() time.Time {
	if len(f.NowTimes) < 1 {
		return time.Now()
	}

	time := f.NowTimes[0]
	if len(f.NowTimes) > 0 {
		f.NowTimes = f.NowTimes[1:]
	}
	return time
}
