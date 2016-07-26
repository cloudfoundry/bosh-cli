package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	fakecmdconf "github.com/cloudfoundry/bosh-init/cmd/config/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("TargetsCmd", func() {
	var (
		config  *fakecmdconf.FakeConfig
		ui      *fakeui.FakeUI
		command EnvironmentsCmd
	)

	BeforeEach(func() {
		config = &fakecmdconf.FakeConfig{}
		ui = &fakeui.FakeUI{}
		command = NewEnvironmentsCmd(config, ui)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("lists targets", func() {
			config.EnvironmentsReturns([]cmdconf.Environment{
				{Alias: "target1-alias", URL: "target1-url"},
				{Alias: "target2-alias", URL: "target2-url"},
			})

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "targets",

				Header: []string{"URL", "Alias"},

				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("target1-url"),
						boshtbl.NewValueString("target1-alias"),
					},
					{
						boshtbl.NewValueString("target2-url"),
						boshtbl.NewValueString("target2-alias"),
					},
				},
			}))
		})
	})
})
