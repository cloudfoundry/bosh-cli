package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("VariablesCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    VariablesCmd
		opts       VariablesOpts
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewVariablesCmd(ui, deployment)
		opts = VariablesOpts{}
	})

	Describe("Run", func() {
		act := func() error { return command.Run(opts) }

		It("lists variables", func() {
			variables := []boshdir.VariableResult{
				{ID: "1", Name: "foo-1", Type: "password"},
				{ID: "2", Name: "foo-2", Type: ""},
				{ID: "3", Name: "foo-3", Type: "certificate"},
				{ID: "4", Name: "foo-4", Type: ""},
			}
			deployment.VariablesReturns(variables, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "variables",

				Header: []boshtbl.Header{boshtbl.NewHeader("ID"), boshtbl.NewHeader("Name"), boshtbl.NewHeader("Type")},

				SortBy: []boshtbl.ColumnSort{
					{Column: 1, Asc: true},
				},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.NewValueString("1"),
						boshtbl.NewValueString("foo-1"),
						boshtbl.NewValueString("password"),
					},
					{
						boshtbl.NewValueString("2"),
						boshtbl.NewValueString("foo-2"),
						boshtbl.NewValueString(""),
					},
					{
						boshtbl.NewValueString("3"),
						boshtbl.NewValueString("foo-3"),
						boshtbl.NewValueString("certificate"),
					},
					{
						boshtbl.NewValueString("4"),
						boshtbl.NewValueString("foo-4"),
						boshtbl.NewValueString(""),
					},
				},
			}))
		})

		It("returns error if getting variables fails", func() {
			deployment.VariablesReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		XContext("when type flag is specified", func() {
			Context("type is certificate", func() {

				BeforeEach(func() {
					opts = VariablesOpts{Type: "certificate"}
				})

				It("should return only the type specified", func() {
					variables := []boshdir.VariableCertResult{
						{ID: "1", Name: "foo-1", ExpiryDate: "2019-11-30T15:51:28Z", DaysLeft: 364},
						{ID: "2", Name: "foo-2", ExpiryDate: "2018-11-30T15:51:28Z", DaysLeft: 0},
						{ID: "3", Name: "foo-3", ExpiryDate: "2017-11-30T15:51:28Z", DaysLeft: 20},
						{ID: "4", Name: "foo-4", ExpiryDate: "2016-11-30T15:51:28Z", DaysLeft: -364},
					}
					deployment.VariableCertsReturns(variables, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Content: "variables",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("ID"),
							boshtbl.NewHeader("Name"),
							boshtbl.NewHeader("Expiry Date"),
							boshtbl.NewHeader("Days Left"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 1, Asc: true},
						},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("1"),
								boshtbl.NewValueString("foo-1"),
								boshtbl.NewValueString("2019-11-30T15:51:28Z"),
								boshtbl.NewValueInt(364),
							},
							{
								boshtbl.NewValueString("2"),
								boshtbl.NewValueString("foo-2"),
								boshtbl.NewValueString("2018-11-30T15:51:28Z"),
								boshtbl.NewValueInt(0),
							},
							{
								boshtbl.NewValueString("3"),
								boshtbl.NewValueString("foo-3"),
								boshtbl.NewValueString("2017-11-30T15:51:28Z"),
								boshtbl.NewValueInt(20),
							},
							{
								boshtbl.NewValueString("4"),
								boshtbl.NewValueString("foo-4"),
								boshtbl.NewValueString("2016-11-30T15:51:28Z"),
								boshtbl.NewValueInt(-364),
							},
						},
					}))

				})
			})

			XContext("when type is not 'certificate", func() {
				unSupportedType := "not-supported-type"

				BeforeEach(func() {
					opts = VariablesOpts{Type: unSupportedType}
				})

				It("should raise error", func() {
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("type: '%s' not supported", unSupportedType)))
				})

			})
		})
	})
})
