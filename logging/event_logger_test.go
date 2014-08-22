package logging_test

import (
	"errors"
	"fmt"
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/logging"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	faketime "github.com/cloudfoundry/bosh-micro-cli/time/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("EventLogger", func() {
	var (
		eventLogger EventLogger
		ui          *fakeui.FakeUI
		eventName   string
		timeService *faketime.FakeService
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		timeService = &faketime.FakeService{}
		eventLogger = NewEventLogger(ui, timeService)
	})

	Context("when given an event name and a func", func() {
		BeforeEach(func() {
			eventName = "compiling packages"
			now := time.Now()
			then := now.Add(3723 * time.Second)
			timeService.NowTimes = []time.Time{now, then}
		})

		It("sends the event to the ui", func() {
			eventLogger.TrackAndLog(eventName, func() error {
				return nil
			})
			Expect(ui.Said).To(ContainElement(fmt.Sprintf("Started %s.", eventName)))
		})

		It("executes the function", func() {
			funcCalled := false
			eventFunc := func() error {
				funcCalled = true
				return nil
			}
			err := eventLogger.TrackAndLog(eventName, eventFunc)
			Expect(err).ToNot(HaveOccurred())
			Expect(funcCalled).To(BeTrue())
		})

		Context("when the function returns nil", func() {
			It("prints Done with the duration of the function call", func() {
				eventLogger.TrackAndLog(eventName, func() error {
					return nil
				})
				Expect(ui.Said).To(ContainElement(" Done (01:02:03)"))
			})
		})

		Context("when the function returns an error", func() {
			It("returns the error", func() {
				eventFunc := func() error {
					return errors.New("There has been an error")
				}
				err := eventLogger.TrackAndLog(eventName, eventFunc)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("when given an empty name and a function", func() {
		BeforeEach(func() {
			eventName = ""
			now := time.Now()
			then := now.Add(3723 * time.Second)
			timeService.NowTimes = []time.Time{now, then}
		})

		It("returns an error", func() {
			err := eventLogger.TrackAndLog(eventName, func() error {
				return nil
			})
			Expect(err).To(HaveOccurred())
		})
	})
})
