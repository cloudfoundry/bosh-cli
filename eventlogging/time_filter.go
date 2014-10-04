package eventlogging

import (
	boshtime "github.com/cloudfoundry/bosh-agent/time"
)

type timeFilter struct {
	timeService boshtime.Service
}

func NewTimeFilter(timeService boshtime.Service) EventFilter {
	return &timeFilter{
		timeService: timeService,
	}
}

func (f *timeFilter) Filter(event *Event) error {
	event.Time = f.timeService.Now()
	return nil
}
