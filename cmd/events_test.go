package cmd_test

import (
	//"errors"
	//"time"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	//boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
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

		//act := func() error { return command.Run(opts) }

		Context("when events are requested", func() {
			It("lists events", func() {
				events := []boshdir.Event{
					//&fakedir.FakeTask{
					//	IDStub: func() int { return 4 },
					//	CreatedAtStub: func() time.Time {
					//		return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
					//	},
					//
					//	StateStub: func() string { return "state" },
					//	UserStub:  func() string { return "user" },
					//
					//	DescriptionStub: func() string { return "description" },
					//	ResultStub:      func() string { return "result" },
					//},
					//&fakedir.FakeTask{
					//	IDStub: func() int { return 5 },
					//	CreatedAtStub: func() time.Time {
					//		return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
					//	},
					//
					//	StateStub:   func() string { return "error" },
					//	IsErrorStub: func() bool { return true },
					//	UserStub:    func() string { return "user2" },
					//
					//	DescriptionStub: func() string { return "description2" },
					//	ResultStub:      func() string { return "result2" },
					//},
				}

				director.EventsReturns(events, nil)
				//
				//err := act()
				//Expect(err).ToNot(HaveOccurred())
				//
				//Expect(ui.Table).To(Equal(boshtbl.Table{
				//	Content: "tasks",
				//
				//	Header: []string{"#", "State", "Created At", "User", "Description", "Result"},
				//
				//	SortBy: []boshtbl.ColumnSort{{Column: 0}},
				//
				//	Rows: [][]boshtbl.Value{
				//		{
				//			boshtbl.NewValueInt(4),
				//			boshtbl.ValueFmt{V: boshtbl.NewValueString("state"), Error: false},
				//			boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
				//			boshtbl.NewValueString("user"),
				//			boshtbl.NewValueString("description"),
				//			boshtbl.NewValueString("result"),
				//		},
				//		{
				//			boshtbl.NewValueInt(5),
				//			boshtbl.ValueFmt{V: boshtbl.NewValueString("error"), Error: true},
				//			boshtbl.NewValueTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)),
				//			boshtbl.NewValueString("user2"),
				//			boshtbl.NewValueString("description2"),
				//			boshtbl.NewValueString("result2"),
				//		},
				//	},
				//}))
			})

			// TODO:
			It("filters events based on 'before-id' option", func() {})
			It("filters events based on 'before' option", func() {})
			It("filters events based on 'after' option", func() {})
			It("filters events based on 'deployment' option", func() {})
			It("filters events based on 'task' option", func() {})
			It("filters events based on 'instance' option", func() {})

			//It("returns error if events cannot be retrieved", func() {
			//	director.EventsReturns(nil, errors.New("fake-err"))
			//
			//	err := act()
			//	Expect(err).To(HaveOccurred())
			//	Expect(err.Error()).To(ContainSubstring("fake-err"))
			//})
		})

	})
})
