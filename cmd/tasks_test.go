package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("TasksCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.TasksCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewTasksCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			tasksOpts opts.TasksOpts
		)

		BeforeEach(func() {
			tasksOpts = opts.TasksOpts{}
		})

		act := func() error { return command.Run(tasksOpts) }

		Context("when current tasks are requested", func() {
			It("lists current tasks", func() {
				tasks := []boshdir.Task{
					&fakedir.FakeTask{
						IDStub: func() int { return 4 },
						StartedAtStub: func() time.Time {
							return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
						},
						FinishedAtStub: func() time.Time {
							return time.Unix(0, 0)
						},

						StateStub:          func() string { return "processing" },
						UserStub:           func() string { return "user" },
						DeploymentNameStub: func() string { return "deployment" },

						DescriptionStub: func() string { return "description" },
						ResultStub:      func() string { return "result" },
					},
					&fakedir.FakeTask{
						IDStub: func() int { return 5 },
						StartedAtStub: func() time.Time {
							return time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC)
						},
						FinishedAtStub: func() time.Time {
							return time.Date(2012, time.December, 10, 23, 0, 0, 0, time.UTC)
						},

						StateStub:          func() string { return "error" },
						IsErrorStub:        func() bool { return true },
						UserStub:           func() string { return "user2" },
						DeploymentNameStub: func() string { return "deployment2" },

						DescriptionStub: func() string { return "description2" },
						ResultStub:      func() string { return "result2" },
					},
				}

				director.CurrentTasksReturns(tasks, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "tasks",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("State"),
						boshtbl.NewHeader("Started At"),
						boshtbl.NewHeader("Finished At"),
						boshtbl.NewHeader("User"),
						boshtbl.NewHeader("Deployment"),
						boshtbl.NewHeader("Description"),
						boshtbl.NewHeader("Result"),
					},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.ValueFmt{V: boshtbl.NewValueString("processing"), Error: false},
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("-"),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("description"),
							boshtbl.NewValueString("result"),
						},
						{
							boshtbl.NewValueInt(5),
							boshtbl.ValueFmt{V: boshtbl.NewValueString("error"), Error: true},
							boshtbl.NewValueTime(time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueTime(time.Date(2012, time.December, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user2"),
							boshtbl.NewValueString("deployment2"),
							boshtbl.NewValueString("description2"),
							boshtbl.NewValueString("result2"),
						},
					},
				}))
			})

			It("filters tasks based options", func() {
				director.CurrentTasksReturns(nil, nil)

				tasksOpts = opts.TasksOpts{}

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(director.CurrentTasksArgsForCall(0)).To(Equal(boshdir.TasksFilter{
					All: true,
				}))

				tasksOpts.All = true
				tasksOpts.Deployment = "deployment"

				err = act()
				Expect(err).ToNot(HaveOccurred())
				Expect(director.CurrentTasksArgsForCall(1)).To(Equal(boshdir.TasksFilter{
					All:        true,
					Deployment: "deployment",
				}))
			})

			It("returns error if tasks cannot be retrieved", func() {
				director.CurrentTasksReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when recent tasks are requested", func() {
			BeforeEach(func() {
				recent := 30
				tasksOpts.Recent = &recent
			})

			It("lists recent tasks", func() {
				tasks := []boshdir.Task{
					&fakedir.FakeTask{
						IDStub: func() int { return 4 },
						StartedAtStub: func() time.Time {
							return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
						},
						FinishedAtStub: func() time.Time {
							return time.Date(2009, time.December, 10, 23, 0, 0, 0, time.UTC)
						},

						StateStub:          func() string { return "state" },
						UserStub:           func() string { return "user" },
						DeploymentNameStub: func() string { return "deployment" },

						DescriptionStub: func() string { return "description" },
						ResultStub:      func() string { return "result" },
					},
					&fakedir.FakeTask{
						IDStub: func() int { return 5 },
						StartedAtStub: func() time.Time {
							return time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC)
						},
						FinishedAtStub: func() time.Time {
							return time.Date(2012, time.December, 10, 23, 0, 0, 0, time.UTC)
						},

						StateStub:          func() string { return "error" },
						IsErrorStub:        func() bool { return true },
						UserStub:           func() string { return "user2" },
						DeploymentNameStub: func() string { return "deployment2" },

						DescriptionStub: func() string { return "description2" },
						ResultStub:      func() string { return "result2" },
					},
				}

				director.RecentTasksReturns(tasks, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "tasks",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("State"),
						boshtbl.NewHeader("Started At"),
						boshtbl.NewHeader("Finished At"),
						boshtbl.NewHeader("User"),
						boshtbl.NewHeader("Deployment"),
						boshtbl.NewHeader("Description"),
						boshtbl.NewHeader("Result"),
					},

					SortBy: []boshtbl.ColumnSort{{Column: 0}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueInt(4),
							boshtbl.ValueFmt{V: boshtbl.NewValueString("state"), Error: false},
							boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueTime(time.Date(2009, time.December, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("description"),
							boshtbl.NewValueString("result"),
						},
						{
							boshtbl.NewValueInt(5),
							boshtbl.ValueFmt{V: boshtbl.NewValueString("error"), Error: true},
							boshtbl.NewValueTime(time.Date(2012, time.November, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueTime(time.Date(2012, time.December, 10, 23, 0, 0, 0, time.UTC)),
							boshtbl.NewValueString("user2"),
							boshtbl.NewValueString("deployment2"),
							boshtbl.NewValueString("description2"),
							boshtbl.NewValueString("result2"),
						},
					},
				}))
			})

			It("filters tasks based on options", func() {
				director.RecentTasksReturns(nil, nil)

				Expect(act()).ToNot(HaveOccurred())
				_, filter := director.RecentTasksArgsForCall(0)
				Expect(filter).To(Equal(boshdir.TasksFilter{}))

				tasksOpts.All = true
				tasksOpts.Deployment = "deployment"

				Expect(act()).ToNot(HaveOccurred())
				_, filter = director.RecentTasksArgsForCall(1)
				Expect(filter).To(Equal(boshdir.TasksFilter{
					All:        true,
					Deployment: "deployment",
				}))
			})

			It("requests specific number of tasks", func() {
				director.RecentTasksReturns(nil, nil)

				Expect(act()).ToNot(HaveOccurred())
				limit, _ := director.RecentTasksArgsForCall(0)
				Expect(limit).To(Equal(30))
			})

			It("returns error if tasks cannot be retrieved", func() {
				director.RecentTasksReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
