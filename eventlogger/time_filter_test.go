package eventlogger_test

import (
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/eventlogger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"
)

var _ = Describe("TimeFilter", func() {
	var (
		timeService *faketime.FakeService
		filter      EventFilter
	)

	BeforeEach(func() {
		timeService = &faketime.FakeService{}
		filter = NewTimeFilter(timeService)
	})

	It("adds the current time to the event", func() {
		event := &Event{
			Stage: "fake-stage",
			Task:  "fake-task",
			State: Started,
		}

		expectedTime := time.Now()
		timeService.NowTimes = []time.Time{expectedTime}
		filter.Filter(event)

		Expect(*event).To(Equal(Event{
			Time:  expectedTime,
			Stage: "fake-stage",
			Task:  "fake-task",
			State: Started,
		}))
	})

})
