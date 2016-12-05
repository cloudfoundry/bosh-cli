package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("VarsCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    VarsCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewVarsCmd(ui, deployment)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("lists config vars", func() {
			configVarsResults := []boshdir.ConfigVarsResult{
				{PlaceholderID: "1", PlaceholderName: "foo-1"},
				{PlaceholderID: "2", PlaceholderName: "foo-2"},
				{PlaceholderID: "3", PlaceholderName: "foo-3"},
				{PlaceholderID: "4", PlaceholderName: "foo-4"},
			}
			deployment.ConfigVarsReturns(configVarsResults, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table).To(Equal(boshtbl.Table{
				//Content: "vars",

				Header: []string{"ID", "Name"},

				SortBy: []boshtbl.ColumnSort{
					{Column: 0, Asc: true},
					{Column: 1},
				},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("1"),
						boshtbl.NewValueString("foo-1"),
					},
					{
						boshtbl.NewValueString("2"),
						boshtbl.NewValueString("foo-2"),
					},
					{
						boshtbl.NewValueString("3"),
						boshtbl.NewValueString("foo-3"),
					},
					{
						boshtbl.NewValueString("4"),
						boshtbl.NewValueString("foo-4"),
					},
				},
			}))
		})

		It("returns error if getting config vars fails", func() {
			deployment.ConfigVarsReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
