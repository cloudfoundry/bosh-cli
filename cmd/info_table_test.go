package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("InfoTable", func() {
	var (
		ui *fakeui.FakeUI
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
	})

	Describe("Print", func() {
		It("shows full information about target", func() {
			info := boshdir.Info{
				Name:    "director-name",
				UUID:    "director-uuid",
				Version: "director-version",

				User: "user",

				Features: map[string]bool{
					"snapshots":     true,
					"compiled_pkgs": false,
				},

				CPI: "cpi",
			}

			InfoTable{Info: info, UI: ui}.Print()

			Expect(ui.Table.Header).To(BeEmpty())
			Expect(ui.Table.Rows).To(Equal([][]boshtbl.Value{
				{
					boshtbl.NewValueString("Name"),
					boshtbl.NewValueString("director-name"),
				},
				{
					boshtbl.NewValueString("UUID"),
					boshtbl.NewValueString("director-uuid"),
				},
				{
					boshtbl.NewValueString("Version"),
					boshtbl.NewValueString("director-version"),
				},
				{
					boshtbl.NewValueString("CPI"),
					boshtbl.NewValueString("cpi"),
				},
				{
					boshtbl.NewValueString("Features"),
					boshtbl.NewValueStrings([]string{"compiled_pkgs: disabled", "snapshots: enabled"}),
				},
				{
					boshtbl.NewValueString("User"),
					boshtbl.NewValueString("user"),
				},
			}))
		})

		It("shows partial information about target when not all of it is available", func() {
			info := boshdir.Info{
				Name:    "director-name",
				UUID:    "director-uuid",
				Version: "director-version",
			}

			InfoTable{Info: info, UI: ui}.Print()

			Expect(ui.Table.Header).To(BeEmpty())
			Expect(ui.Table.Rows).To(Equal([][]boshtbl.Value{
				{
					boshtbl.NewValueString("Name"),
					boshtbl.NewValueString("director-name"),
				},
				{
					boshtbl.NewValueString("UUID"),
					boshtbl.NewValueString("director-uuid"),
				},
				{
					boshtbl.NewValueString("Version"),
					boshtbl.NewValueString("director-version"),
				},
				{
					boshtbl.NewValueString("User"),
					boshtbl.NewValueString("(not logged in)"),
				},
			}))
		})
	})
})
