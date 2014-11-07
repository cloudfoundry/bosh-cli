package eventlogger_test

import (
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	fakelog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

var _ = Describe("EventLogger", func() {
	var (
		eventLogger EventLogger
		ui          bmui.UI

		uiOut *bytes.Buffer
		uiErr *bytes.Buffer
	)

	BeforeEach(func() {
		uiOut = bytes.NewBufferString("")
		uiErr = bytes.NewBufferString("")
		logger := boshlog.NewLogger(boshlog.LevelNone)
		ui = bmui.NewUI(uiOut, uiErr, logger)
		eventLogger = NewEventLogger(ui)
	})

	Describe("AddEvent", func() {
		It("tells the UI to print out start event", func() {
			event := Event{
				Stage: "fake-stage",
				State: Started,
				Task:  "fake-task-1",
			}
			eventLogger.AddEvent(event)
			output := uiOut.String()
			Expect(output).To(ContainSubstring("Started fake-stage > fake-task-1."))
		})

		Context("when filters are configured", func() {
			var (
				filter *fakelog.FakeEventFilter
			)
			BeforeEach(func() {
				filter = fakelog.NewFakeEventFilter()
				filters := []EventFilter{filter}
				eventLogger = NewEventLoggerWithFilters(ui, filters)
			})

			It("modifies the event using the filters", func() {
				filter.SetFilterBehavior(func(event *Event) error {
					event.Stage = "filtered-fake-stage"
					return nil
				})

				event := Event{
					Stage: "fake-stage",
					State: Started,
					Task:  "fake-task-1",
				}
				eventLogger.AddEvent(event)

				output := uiOut.String()
				Expect(output).To(ContainSubstring("Started filtered-fake-stage > fake-task-1."))
			})
		})

		Context("When all the tasks are finished", func() {
			BeforeEach(func() {
				now := time.Now()
				eventLogger.StartStage("fake-stage")
				eventLogger.AddEvent(Event{
					Time:  now,
					Stage: "fake-stage",
					Task:  "fake-task-1",
					State: Started,
				})

				eventLogger.AddEvent(Event{
					Time:  now.Add(1 * time.Second),
					Stage: "fake-stage",
					Task:  "fake-task-1",
					State: Finished,
				})

				eventLogger.AddEvent(Event{
					Time:  now.Add(2 * time.Second),
					Stage: "fake-stage",
					Task:  "fake-task-2",
					State: Started,
				})

				eventLogger.AddEvent(Event{
					Time:  now.Add(3 * time.Second),
					Stage: "fake-stage",
					Task:  "fake-task-2",
					State: Finished,
				})

				eventLogger.FinishStage("fake-stage")
			})

			It("tells the UI to start the stage", func() {
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Done fake-stage\n"))
			})

			It("tells the UI to print out Done when the task is finished", func() {
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Started fake-stage > fake-task-2... done. (00:00:01)\n"))
			})

			It("tells the UI to finish the stage with an extra linebreak", func() {
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Done fake-stage\n\n"))
			})
		})

		Context("when task failed", func() {
			It("tells UI to print out an error message", func() {
				now := time.Now()
				eventLogger.AddEvent(Event{
					Time:  now,
					Stage: "fake-stage",
					Task:  "fake-task-1",
					State: Started,
				})

				eventLogger.AddEvent(Event{
					Time:    now.Add(1 * time.Second),
					Stage:   "fake-stage",
					Task:    "fake-task-1",
					State:   Failed,
					Message: "fake-fail-message",
				})
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Started fake-stage > fake-task-1... failed (fake-fail-message). (00:00:01)\n"))
			})
		})

		Context("when task is skipped", func() {
			It("tells UI to print out a skipped message", func() {
				now := time.Now()
				eventLogger.AddEvent(Event{
					Time:  now,
					Stage: "fake-stage",
					Task:  "fake-task-1",
					State: Started,
				})

				eventLogger.AddEvent(Event{
					Time:    now.Add(1 * time.Second),
					Stage:   "fake-stage",
					Task:    "fake-task-1",
					State:   Skipped,
					Message: "fake-skipped-message",
				})
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Started fake-stage > fake-task-1... skipped (fake-skipped-message).\n"))
			})
		})

		Context("when a unsupported event state was received", func() {
			It("returns error", func() {
				error := eventLogger.AddEvent(Event{
					State: "fake-state",
				})
				Expect(error).To(HaveOccurred())
				Expect(error.Error()).To(ContainSubstring("Unsupported event state `fake-state'"))
			})
		})
	})
})
