package logging_test

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	. "github.com/cloudfoundry/bosh-micro-cli/logging"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	faketime "github.com/cloudfoundry/bosh-micro-cli/time/fakes"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

var _ = Describe("EventLogger", func() {
	var (
		eventLogger EventLogger
		ui          bmui.UI
		eventName   string
		timeService *faketime.FakeService
		stdout      *bytes.Buffer
		stderr      *bytes.Buffer
	)

	BeforeEach(func() {
		stdout = bytes.NewBufferString("")
		stderr = bytes.NewBufferString("")
		ui = bmui.NewUI(stdout, stderr)
		timeService = &faketime.FakeService{}
		eventLogger = NewEventLogger(ui, timeService)
	})

	Context("TrackAndLog", func() {
		Context("when given an event name and a func", func() {
			var groupName string

			BeforeEach(func() {
				groupName = "compiling packages"
				eventLogger.StartGroup(groupName)
				eventName = "ruby/123"
				now := time.Now()
				then := now.Add(3723 * time.Second)
				timeService.NowTimes = []time.Time{now, then}
			})

			It("sends the event to the ui", func() {
				eventLogger.TrackAndLog(eventName, func() error {
					return nil
				})
				Expect(stdout).To(ContainSubstring(fmt.Sprintf("%s > %s", groupName, eventName)))
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
					Expect(stdout).To(ContainSubstring(" Done (01:02:03)\n"))
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

	Context("StartGroup", func() {
		Context("when given an empty string", func() {
			It("returns an error", func() {
				group := ""
				err := eventLogger.StartGroup(group)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when given a non-empty string", func() {
			It("tells the ui to print the string", func() {
				group := "compiling packages"
				err := eventLogger.StartGroup(group)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring(fmt.Sprintf("Started %s\n", group)))
			})
		})
	})

	Context("FinishGroup", func() {
		Context("when a group was not previously started", func() {
			It("returns an error", func() {
				err := eventLogger.FinishGroup()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a group was previously started", func() {
			It("tells the ui the group is done", func() {
				group := "compiling packages"
				err := eventLogger.StartGroup(group)
				Expect(err).NotTo(HaveOccurred())
				err = eventLogger.FinishGroup()
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(ContainSubstring(fmt.Sprintf("Done %s\n", group)))
			})

			Context("when the group is finished twice", func() {
				It("returns an error", func() {
					group := "compiling packages"
					err := eventLogger.StartGroup(group)
					Expect(err).NotTo(HaveOccurred())
					err = eventLogger.FinishGroup()
					Expect(err).NotTo(HaveOccurred())
					err = eventLogger.FinishGroup()
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
