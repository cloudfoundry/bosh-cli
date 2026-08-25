package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	cmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config"
	fakecmdconf "github.com/cloudfoundry/bosh-cli/v7/cmd/config/configfakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("EnvironmentsCmd", func() {
	var (
		config  *fakecmdconf.FakeConfig
		testUI  *testui.Ui
		command cmd.EnvironmentsCmd
	)

	BeforeEach(func() {
		config = &fakecmdconf.FakeConfig{}
		testUI = &testui.Ui{}
		command = cmd.NewEnvironmentsCmd(config, testUI)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("lists environments", func() {
			config.EnvironmentsReturns([]cmdconf.Environment{
				{Alias: "environment1-alias", URL: "environment1-url"},
				{Alias: "environment2-alias", URL: "environment2-url"},
			})

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Table).To(Equal(boshtbl.Table{
				Content: "environments",

				Header: []boshtbl.Header{
					boshtbl.NewHeader("URL"),
					boshtbl.NewHeader("Alias"),
				},

				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("environment1-url"),
						boshtbl.NewValueString("environment1-alias"),
					},
					{
						boshtbl.NewValueString("environment2-url"),
						boshtbl.NewValueString("environment2-alias"),
					},
				},
			}))
		})
	})
})
