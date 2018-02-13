package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("ConfigTable", func() {
	var (
		ui *fakeui.FakeUI
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
	})

	Describe("Print", func() {
		It("shows config info as transposed tabled", func() {
			config := boshdir.Config{ID: "123", Type: "my-type", Name: "my-name", CreatedAt: "sunday", Content: "some-content"}
			ConfigTable{Config: config, UI: ui}.Print()
			Expect(ui.Table).To(Equal(
				boshtbl.Table{
					Content: "config",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("ID"),
						boshtbl.NewHeader("Type"),
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("Created At"),
						boshtbl.NewHeader("Content"),
					},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("123"),
							boshtbl.NewValueString("my-type"),
							boshtbl.NewValueString("my-name"),
							boshtbl.NewValueString("sunday"),
							boshtbl.NewValueString("some-content"),
						},
					},

					Notes: []string{},

					FillFirstColumn: true,

					Transpose: true,
				}))
		})
	})
})
