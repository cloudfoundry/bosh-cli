package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("EventsCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  EventsCmd
		events   []boshdir.Event
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewEventsCmd(ui, director)
		events = []boshdir.Event{
			&fakedir.FakeEvent{
				IDStub:        func() string { return "4" },
				ParentIDStub:  func() string { return "1" },
				TimestampStub: func() time.Time { return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC) },

				UserStub: func() string { return "user" },

				ActionStub:         func() string { return "action" },
				ObjectTypeStub:     func() string { return "object_type" },
				ObjectNameStub:     func() string { return "object_name" },
				TaskIDStub:         func() string { return "taskID" },
				DeploymentNameStub: func() string { return "deploymentName" },
				InstanceStub:       func() string { return "instance" },
				ContextStub:        func() map[string]interface{} { return map[string]interface{}{"user": "bosh_z$"} },
			},
			&fakedir.FakeEvent{
				IDStub:        func() string { return "5" },
				TimestampStub: func() time.Time { return time.Date(2090, time.November, 10, 23, 0, 0, 0, time.UTC) },

				UserStub: func() string { return "user2" },

				ActionStub:         func() string { return "action2" },
				ObjectTypeStub:     func() string { return "object_type2" },
				ObjectNameStub:     func() string { return "object_name2" },
				TaskIDStub:         func() string { return "task2" },
				DeploymentNameStub: func() string { return "deployment2" },
				InstanceStub:       func() string { return "instance2" },
				ContextStub:        func() map[string]interface{} { return make(map[string]interface{}) },
			},
		}
	})

	Describe("Run", func() {
		var (
			opts EventsOpts
		)

		It("lists events", func() {
			director.EventsReturns(events, nil)

			err := command.Run(opts)
			Expect(err).ToNot(HaveOccurred())

			Expect(director.EventsArgsForCall(0)).To(Equal(boshdir.EventsFilter{}))

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "events",

				Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task ID", "Deployment", "Instance", "Context"},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("4 <- 1"),
						boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
						boshtbl.NewValueString("user"),
						boshtbl.NewValueString("action"),
						boshtbl.NewValueString("object_type"),
						boshtbl.NewValueString("object_name"),
						boshtbl.NewValueString("taskID"),
						boshtbl.NewValueString("deploymentName"),
						boshtbl.NewValueString("instance"),
						boshtbl.NewValueInterface(map[string]interface{}{"user": "bosh_z$"}),
					},
					{
						boshtbl.NewValueString("5"),
						boshtbl.NewValueTime(time.Date(2090, time.November, 10, 23, 0, 0, 0, time.UTC)),
						boshtbl.NewValueString("user2"),
						boshtbl.NewValueString("action2"),
						boshtbl.NewValueString("object_type2"),
						boshtbl.NewValueString("object_name2"),
						boshtbl.NewValueString("task2"),
						boshtbl.NewValueString("deployment2"),
						boshtbl.NewValueString("instance2"),
						boshtbl.NewValueInterface(map[string]interface{}{}),
					},
				},
			}))
		})

		It("filters events based on options", func() {
			beforeID := "0"
			before := time.Date(2050, time.November, 10, 23, 0, 0, 0, time.UTC).String()
			after := time.Date(3055, time.November, 10, 23, 0, 0, 0, time.UTC).String()
			deploymentName := "deploymentName"
			taskID := "task"
			instance := "instance2"
			opts.BeforeID = &beforeID
			opts.Before = &before
			opts.After = &after
			opts.DeploymentName = &deploymentName
			opts.TaskID = &taskID
			opts.Instance = &instance

			director.EventsReturns(nil, nil)

			err := command.Run(opts)
			Expect(err).ToNot(HaveOccurred())

			Expect(director.EventsArgsForCall(0)).To(Equal(boshdir.EventsFilter{
				BeforeID:       &beforeID,
				Before:         &before,
				After:          &after,
				DeploymentName: &deploymentName,
				TaskID:         &taskID,
				Instance:       &instance,
			}))
		})

		It("returns error if events cannot be retrieved", func() {
			director.EventsReturns(nil, errors.New("fake-err"))

			err := command.Run(opts)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
