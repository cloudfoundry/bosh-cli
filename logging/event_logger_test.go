package logging_test

import (
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"

	. "github.com/cloudfoundry/bosh-micro-cli/logging"
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
				Total: 2,
				State: "started",
				Task:  "fake-task-1",
				Index: 1,
			}
			eventLogger.AddEvent(event)
			output := uiOut.String()
			Expect(output).To(ContainSubstring("Started fake-stage\n"))
			Expect(output).To(ContainSubstring("Started fake-stage > fake-task-1."))
		})

		Context("When all the tasks are finished", func() {
			BeforeEach(func() {
				now := time.Now()
				eventLogger.AddEvent(Event{
					Time:  now,
					Stage: "fake-stage",
					Total: 2,
					Task:  "fake-task-1",
					State: "started",
					Index: 1,
				})

				eventLogger.AddEvent(Event{
					Time:  now.Add(1 * time.Second),
					Stage: "fake-stage",
					Total: 2,
					Task:  "fake-task-1",
					State: "finished",
					Index: 1,
				})

				eventLogger.AddEvent(Event{
					Time:  now.Add(2 * time.Second),
					Stage: "fake-stage",
					Total: 2,
					Task:  "fake-task-2",
					State: "started",
					Index: 2,
				})
				eventLogger.AddEvent(Event{
					Time:  now.Add(3 * time.Second),
					Stage: "fake-stage",
					Total: 2,
					Task:  "fake-task-2",
					State: "finished",
					Index: 2,
				})
			})

			It("tells the UI to print out Done when the task is finished", func() {
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Started fake-stage > fake-task-2. Done (00:00:01)\n"))
			})

			It("tells the UI to finish the stage", func() {
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Done fake-stage\n"))
			})
		})

		Context("when task failed", func() {
			It("tells UI to print out an error message", func() {
				now := time.Now()
				eventLogger.AddEvent(Event{
					Time:  now,
					Stage: "fake-stage",
					Total: 2,
					Task:  "fake-task-1",
					State: "started",
					Index: 1,
				})

				eventLogger.AddEvent(Event{
					Time:    now.Add(1 * time.Second),
					Stage:   "fake-stage",
					Total:   2,
					Task:    "fake-task-1",
					State:   "failed",
					Index:   1,
					Message: "fake-fail-message",
				})
				output := uiOut.String()
				Expect(output).To(ContainSubstring("Started fake-stage > fake-task-1. Failed 'fake-fail-message' (00:00:01)\n"))
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
