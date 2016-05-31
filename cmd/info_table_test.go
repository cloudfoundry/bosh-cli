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
					boshtbl.ValueString{"Name"},
					boshtbl.ValueString{"director-name"},
				},
				{
					boshtbl.ValueString{"UUID"},
					boshtbl.ValueString{"director-uuid"},
				},
				{
					boshtbl.ValueString{"Version"},
					boshtbl.ValueString{"director-version"},
				},
				{
					boshtbl.ValueString{"CPI"},
					boshtbl.ValueString{"cpi"},
				},
				{
					boshtbl.ValueString{"Features"},
					boshtbl.ValueStrings{[]string{"compiled_pkgs: disabled", "snapshots: enabled"}},
				},
				{
					boshtbl.ValueString{"User"},
					boshtbl.ValueString{"user"},
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
					boshtbl.ValueString{"Name"},
					boshtbl.ValueString{"director-name"},
				},
				{
					boshtbl.ValueString{"UUID"},
					boshtbl.ValueString{"director-uuid"},
				},
				{
					boshtbl.ValueString{"Version"},
					boshtbl.ValueString{"director-version"},
				},
				{
					boshtbl.ValueString{"User"},
					boshtbl.ValueString{"(not logged in)"},
				},
			}))
		})
	})
})
