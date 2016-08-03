package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
	"time"
)

var _ = Describe("EventsCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  EventsCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewEventsCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts EventsOpts
		)

		BeforeEach(func() {
			opts = EventsOpts{}
		})

		act := func() error { return command.Run(opts) }

		Context("when events are requested", func() {

			events := []boshdir.Event{
				&fakedir.FakeEvent{
					IdStub: func() int {
						return 4
					},
					TimestampStub: func() time.Time {
						return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
					},

					UserStub: func() string {
						return "user"
					},

					ActionStub: func() string {
						return "action"
					},
					ObjectTypeStub: func() string {
						return "object_type"
					},
					ObjectNameStub: func() string {
						return "object_name"
					},
					TaskStub: func() string {
						return "task"
					},
					DeploymentStub: func() string {
						return "deployment"
					},
					InstanceStub: func() string {
						return "instance"
					},
					ContextStub: func() map[string]interface{} {
						ret := make(map[string]interface{})
						ret["user"] = "bosh_z$"
						return ret
					},
				},
				&fakedir.FakeEvent{
					IdStub: func() int {
						return 5
					},
					TimestampStub: func() time.Time {
						return time.Date(2090, time.November, 10, 23, 0, 0, 0, time.UTC)
					},

					UserStub: func() string {
						return "user2"
					},

					ActionStub: func() string {
						return "action2"
					},
					ObjectTypeStub: func() string {
						return "object_type2"
					},
					ObjectNameStub: func() string {
						return "object_name2"
					},
					TaskStub: func() string {
						return "task2"
					},
					DeploymentStub: func() string {
						return "deployment2"
					},
					InstanceStub: func() string {
						return "instance2"
					},
					ContextStub: func() map[string]interface{} {
						ret := make(map[string]interface{})
						return ret
					},
				},
			}

			event4 := []boshdir.Event{
				events[0],
			}

			event5 := []boshdir.Event{
				events[1],
			}

			It("lists events", func() {

				director.EventsReturns(events, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("action"),
							boshtbl.NewValueString("object_type"),
							boshtbl.NewValueString("object_name"),
							boshtbl.NewValueString("task"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance"),
							boshtbl.NewValueString("map[user:bosh_z$]"),
						},
						{
							boshtbl.NewValueInt(5),
							boshtbl.NewValueTime(time.Date(2090, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user2"),
							boshtbl.NewValueString("action2"),
							boshtbl.NewValueString("object_type2"),
							boshtbl.NewValueString("object_name2"),
							boshtbl.NewValueString("task2"),
							boshtbl.NewValueString("deployment2"),
							boshtbl.NewValueString("instance2"),
							boshtbl.NewValueString("map[]"),
						},
					},
				}))

			})

			It("filters events based on 'before-id' option", func() {

				opts.BeforeId = 4

				director.EventsReturns(event4, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("action"),
							boshtbl.NewValueString("object_type"),
							boshtbl.NewValueString("object_name"),
							boshtbl.NewValueString("task"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance"),
							boshtbl.NewValueString("map[user:bosh_z$]"),
						},
					},
				}))
			})
			It("filters events based on 'before' option", func() {
				opts.Before = time.Date(2050, time.November, 10, 23, 0, 0, 0, time.UTC)

				director.EventsReturns(event4, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("action"),
							boshtbl.NewValueString("object_type"),
							boshtbl.NewValueString("object_name"),
							boshtbl.NewValueString("task"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance"),
							boshtbl.NewValueString("map[user:bosh_z$]"),
						},
					},
				}))
			})
			It("filters events based on 'after' option", func() {
				opts.After = time.Date(2050, time.November, 10, 23, 0, 0, 0, time.UTC)

				director.EventsReturns(event5, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(5),
							boshtbl.NewValueTime(time.Date(2090, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user2"),
							boshtbl.NewValueString("action2"),
							boshtbl.NewValueString("object_type2"),
							boshtbl.NewValueString("object_name2"),
							boshtbl.NewValueString("task2"),
							boshtbl.NewValueString("deployment2"),
							boshtbl.NewValueString("instance2"),
							boshtbl.NewValueString("map[]"),
						},
					},
				}))
			})
			It("filters events based on 'deployment' option", func() {

				opts.Deployment = "deployment"

				director.EventsReturns(event4, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("action"),
							boshtbl.NewValueString("object_type"),
							boshtbl.NewValueString("object_name"),
							boshtbl.NewValueString("task"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance"),
							boshtbl.NewValueString("map[user:bosh_z$]"),
						},
					},
				}))
			})
			It("filters events based on 'task' option", func() {
				opts.Task = "task"

				director.EventsReturns(event4, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("action"),
							boshtbl.NewValueString("object_type"),
							boshtbl.NewValueString("object_name"),
							boshtbl.NewValueString("task"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance"),
							boshtbl.NewValueString("map[user:bosh_z$]"),
						},
					},
				}))
			})
			It("filters events based on 'instance' option", func() {
				opts.Instance = "instance2"

				director.EventsReturns(event5, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "events",

					Header: []string{"ID", "Time", "User", "Action", "Object Type", "Object ID", "Task", "Deployment", "Instance", "Context"},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(5),
							boshtbl.NewValueTime(time.Date(2090, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user2"),
							boshtbl.NewValueString("action2"),
							boshtbl.NewValueString("object_type2"),
							boshtbl.NewValueString("object_name2"),
							boshtbl.NewValueString("task2"),
							boshtbl.NewValueString("deployment2"),
							boshtbl.NewValueString("instance2"),
							boshtbl.NewValueString("map[]"),
						},
					},
				}))
			})

			It("returns error if events cannot be retrieved", func() {
				director.EventsReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
