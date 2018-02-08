package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("ConfigsCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  ConfigsCmd
		configs  []boshdir.Config
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewConfigsCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts ConfigsOpts
		)

		BeforeEach(func() {
			opts = ConfigsOpts{Recent: 1}
			configs = []boshdir.Config{
				boshdir.Config{Type: "my-type", Name: "some-name", Team: "team1"},
				boshdir.Config{Type: "my-type", Name: "other-name"},
			}
		})

		act := func() error { return command.Run(opts) }

		It("lists configs", func() {
			director.ListConfigsReturns(configs, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(director.ListConfigsCallCount()).To(Equal(1))
			limit, filter := director.ListConfigsArgsForCall(0)
			Expect(limit).To(Equal(1))
			Expect(filter).To(Equal(boshdir.ConfigsFilter{}))

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "configs",

				Header: []boshtbl.Header{
					boshtbl.NewHeader("ID"),
					boshtbl.NewHeader("Type"),
					boshtbl.NewHeader("Name"),
					boshtbl.NewHeader("Team"),
					boshtbl.NewHeader("Created At"),
				},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString(""),
						boshtbl.NewValueString("my-type"),
						boshtbl.NewValueString("some-name"),
						boshtbl.NewValueString("team1"),
						boshtbl.NewValueString(""),
					},
					{
						boshtbl.NewValueString(""),
						boshtbl.NewValueString("my-type"),
						boshtbl.NewValueString("other-name"),
						boshtbl.NewValueString(""),
						boshtbl.NewValueString(""),
					},
				},
			}))
		})

		It("returns error if configs cannot be listed", func() {
			director.ListConfigsReturns([]boshdir.Config{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		Context("When filtering for type", func() {
			BeforeEach(func() {
				opts = ConfigsOpts{
					Type:   "my-type",
					Recent: 1,
				}
				configs = []boshdir.Config{boshdir.Config{Type: "my-type", Name: "some-name"}}
			})

			It("applies filters for just type", func() {
				director.ListConfigsReturns(configs, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(director.ListConfigsCallCount()).To(Equal(1))
				limit, filter := director.ListConfigsArgsForCall(0)
				Expect(limit).To(Equal(1))
				Expect(filter).To(Equal(boshdir.ConfigsFilter{Type: "my-type"}))

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "configs",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("Type"),
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Team"),
						boshtbl.NewHeader("Created At"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString(""),
							boshtbl.NewValueString("my-type"),
							boshtbl.NewValueString("some-name"),
							boshtbl.NewValueString(""),
							boshtbl.NewValueString(""),
						},
					},
				}))
			})
		})

		Context("When filtering for name", func() {
			BeforeEach(func() {
				opts = ConfigsOpts{
					Name:   "some-name",
					Recent: 1,
				}
				configs = []boshdir.Config{boshdir.Config{Type: "my-type", Name: "some-name"}}
			})

			It("applies filters for just name", func() {
				director.ListConfigsReturns(configs, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(director.ListConfigsCallCount()).To(Equal(1))
				limit, filter := director.ListConfigsArgsForCall(0)
				Expect(limit).To(Equal(1))
				Expect(filter).To(Equal(boshdir.ConfigsFilter{Name: "some-name"}))

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "configs",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("Type"),
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Team"),
						boshtbl.NewHeader("Created At"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString(""),
							boshtbl.NewValueString("my-type"),
							boshtbl.NewValueString("some-name"),
							boshtbl.NewValueString(""),
							boshtbl.NewValueString(""),
						},
					},
				}))
			})
		})

		Context("When filtering for both, type and name", func() {
			BeforeEach(func() {
				opts = ConfigsOpts{
					Type:   "my-type",
					Name:   "some-name",
					Recent: 1,
				}
				configs = []boshdir.Config{boshdir.Config{Type: "my-type", Name: "some-name"}}
			})

			It("applies filters for type and name", func() {
				director.ListConfigsReturns(configs, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(director.ListConfigsCallCount()).To(Equal(1))
				limit, filter := director.ListConfigsArgsForCall(0)
				Expect(limit).To(Equal(1))
				Expect(filter).To(Equal(boshdir.ConfigsFilter{Name: "some-name", Type: "my-type"}))

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "configs",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("Type"),
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Team"),
						boshtbl.NewHeader("Created At"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString(""),
							boshtbl.NewValueString("my-type"),
							boshtbl.NewValueString("some-name"),
							boshtbl.NewValueString(""),
							boshtbl.NewValueString(""),
						},
					},
				}))
			})
		})

		Context("limit is specified", func() {
			BeforeEach(func() {
				opts = ConfigsOpts{Recent: 2}
				configs = []boshdir.Config{
					boshdir.Config{Type: "my-type", Name: "some-name", ID: "123"},
					boshdir.Config{Type: "my-type", Name: "some-name", ID: "234"},
				}
			})

			It("lists outdated configs versioned by ID", func() {
				director.ListConfigsReturns(configs, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(director.ListConfigsCallCount()).To(Equal(1))
				limit, filter := director.ListConfigsArgsForCall(0)
				Expect(limit).To(Equal(2))
				Expect(filter).To(Equal(boshdir.ConfigsFilter{}))

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "configs",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("Type"),
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Team"),
						boshtbl.NewHeader("Created At"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("123"),
							boshtbl.NewValueString("my-type"),
							boshtbl.NewValueString("some-name"),
							boshtbl.NewValueString(""),
							boshtbl.NewValueString(""),
						},
						{
							boshtbl.NewValueString("234"),
							boshtbl.NewValueString("my-type"),
							boshtbl.NewValueString("some-name"),
							boshtbl.NewValueString(""),
							boshtbl.NewValueString(""),
						},
					},
				}))
			})
		})
	})
})
