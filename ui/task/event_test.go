package task_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshuit "github.com/cloudfoundry/bosh-init/ui/task"
)

var _ = Describe("Event", func() {
	Describe("IsSame", func() {
		It("returns false if stage is different", func() {
			e1 := boshuit.Event{Stage: "stage1"}
			e2 := boshuit.Event{Stage: "stage2"}
			Expect(e1.IsSame(e2)).To(BeFalse())
		})

		It("returns false if stage is same but tags are different", func() {
			e1 := boshuit.Event{Stage: "stage", Tags: []string{"tag1"}}
			e2 := boshuit.Event{Stage: "stage", Tags: []string{"tag1", "tag2"}}
			Expect(e1.IsSame(e2)).To(BeFalse())
		})

		It("returns false if stage and tags are same but task is different", func() {
			e1 := boshuit.Event{Stage: "stage", Task: "task1", Tags: []string{"tag"}}
			e2 := boshuit.Event{Stage: "stage", Task: "task2", Tags: []string{"tag"}}
			Expect(e1.IsSame(e2)).To(BeFalse())
		})

		It("returns true if stage is same and tags are empty", func() {
			e1 := boshuit.Event{Stage: "stage", Task: "task"}
			e2 := boshuit.Event{Stage: "stage", Task: "task"}
			Expect(e1.IsSame(e2)).To(BeTrue())
		})

		It("returns true if stage tags, task are same", func() {
			e1 := boshuit.Event{Stage: "stage", Task: "task", Tags: []string{"tag1", "tag2"}}
			e2 := boshuit.Event{Stage: "stage", Task: "task", Tags: []string{"tag1", "tag2"}}
			Expect(e1.IsSame(e2)).To(BeTrue())
		})
	})

	Describe("IsSameGroup", func() {
		It("returns false if stage is different", func() {
			e1 := boshuit.Event{Stage: "stage1"}
			e2 := boshuit.Event{Stage: "stage2"}
			Expect(e1.IsSameGroup(e2)).To(BeFalse())
		})

		It("returns false if stage is same but tags are different", func() {
			e1 := boshuit.Event{Stage: "stage", Tags: []string{"tag1"}}
			e2 := boshuit.Event{Stage: "stage", Tags: []string{"tag1", "tag2"}}
			Expect(e1.IsSameGroup(e2)).To(BeFalse())
		})

		It("returns true if stage is same and tags are empty", func() {
			e1 := boshuit.Event{Stage: "stage"}
			e2 := boshuit.Event{Stage: "stage"}
			Expect(e1.IsSameGroup(e2)).To(BeTrue())
		})

		It("returns true if stage and tags are same", func() {
			e1 := boshuit.Event{Stage: "stage", Tags: []string{"tag1", "tag2"}}
			e2 := boshuit.Event{Stage: "stage", Tags: []string{"tag1", "tag2"}}
			Expect(e1.IsSameGroup(e2)).To(BeTrue())
		})
	})

	Describe("Time", func() {
		It("returns formatted time string", func() {
			e := boshuit.Event{UnixTime: 3793593658}
			Expect(e.Time()).To(Equal(time.Date(2090, time.March, 19, 8, 0, 58, 0, time.UTC)))
		})
	})

	Describe("TimeAsStr", func() {
		It("returns formatted time string", func() {
			e := boshuit.Event{UnixTime: 3793593658}
			Expect(e.TimeAsStr()).To(Equal("Sun Mar 19 08:00:58 UTC 2090"))
		})
	})

	Describe("TimeAsHoursStr", func() {
		It("returns formatted hours string", func() {
			e := boshuit.Event{UnixTime: 100}
			Expect(e.TimeAsHoursStr()).To(Equal("00:01:40"))
		})
	})

	Describe("DurationAsStr", func() {
		It("returns formatted duration since given event's time", func() {
			start := boshuit.Event{UnixTime: 100}
			end := boshuit.Event{UnixTime: 200, StartEvent: &start}
			Expect(start.DurationAsStr(end)).To(Equal("00:01:40"))
		})
	})

	Describe("DurationSinceStartAsStr", func() {
		It("returns empty string if does not have a start", func() {
			Expect(boshuit.Event{}.DurationSinceStartAsStr()).To(Equal(""))
		})

		It("returns formatted duration since the start event's time", func() {
			start := boshuit.Event{UnixTime: 100}
			end := boshuit.Event{UnixTime: 200, StartEvent: &start}
			Expect(end.DurationSinceStartAsStr()).To(Equal("00:01:40"))
		})
	})

	Describe("IsWorthKeeping", func() {
		It("returns true if event is not a progress event", func() {
			Expect(boshuit.Event{State: boshuit.EventStateStarted}.IsWorthKeeping()).To(BeTrue())
			Expect(boshuit.Event{State: boshuit.EventStateInProgress}.IsWorthKeeping()).To(BeFalse())
		})
	})
})
