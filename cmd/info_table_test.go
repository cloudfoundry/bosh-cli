package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("InfoTable", func() {
	var (
		ui *fakeui.FakeUI
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
	})

	Describe("Print", func() {
		It("shows full information about environment", func() {
			info := boshdir.Info{
				Name:            "director-name",
				UUID:            "director-uuid",
				StemcellOS:      "-",
				StemcellVersion: "-",
				Version:         "director-version",

				User: "user",

				Features: map[string]bool{
					"snapshots":     true,
					"compiled_pkgs": false,
				},

				CPI: "cpi",
			}

			cmd.InfoTable{Info: info, UI: ui}.Print()

			Expect(ui.Table.Header).To(Equal([]boshtbl.Header{
				boshtbl.NewHeader("Name"),
				boshtbl.NewHeader("UUID"),
				boshtbl.NewHeader("Version"),
				boshtbl.NewHeader("Director Stemcell"),
				boshtbl.NewHeader("CPI"),
				boshtbl.NewHeader("Features"),
				boshtbl.NewHeader("User"),
			}))
			Expect(ui.Table.Rows).To(HaveLen(1))
			Expect(ui.Table.Rows[0]).To(Equal([]boshtbl.Value{
				boshtbl.NewValueString("director-name"),
				boshtbl.NewValueString("director-uuid"),
				boshtbl.NewValueString("director-version"),
				boshtbl.NewValueString("-/-"),
				boshtbl.NewValueString("cpi"),
				boshtbl.NewValueStrings([]string{"compiled_pkgs: disabled", "snapshots: enabled"}),
				boshtbl.NewValueString("user"),
			}))
		})

		It("shows partial information about environment when not all of it is available", func() {
			info := boshdir.Info{
				Name:    "director-name",
				UUID:    "director-uuid",
				Version: "director-version",
			}

			cmd.InfoTable{Info: info, UI: ui}.Print()

			Expect(ui.Table.Header).To(Equal([]boshtbl.Header{
				boshtbl.NewHeader("Name"),
				boshtbl.NewHeader("UUID"),
				boshtbl.NewHeader("Version"),
				boshtbl.NewHeader("User"),
			}))
			Expect(ui.Table.Rows).To(HaveLen(1))
			Expect(ui.Table.Rows[0]).To(Equal([]boshtbl.Value{
				boshtbl.NewValueString("director-name"),
				boshtbl.NewValueString("director-uuid"),
				boshtbl.NewValueString("director-version"),
				boshtbl.NewValueString("(not logged in)"),
			}))
		})
	})
})
