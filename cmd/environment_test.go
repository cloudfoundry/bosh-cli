package cmd_test

import (
	"errors"
	"github.com/fatih/color"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("EnvironmentCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  EnvironmentCmd
		opts     EnvironmentOpts
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewEnvironmentCmd(ui, director)
		opts = EnvironmentOpts{}
	})

	Describe("Run", func() {
		act := func() error { return command.Run(opts) }

		It("outputs a table that should be transposed", func() {
			info := boshdir.Info{}
			director.InfoReturns(info, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table.Transpose).To(Equal(true))
		})

		Context("when all information is present", func() {
			It("outputs a table with columns in the correct order", func() {
				info := boshdir.Info{
					Name:     "director-name",
					UUID:     "director-uuid",
					Version:  "director-version",
					CPI:      "cpi-info",
					Features: map[string]bool{"feature-1": true},
					User:     "best-user",
				}
				director.InfoReturns(info, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("UUID"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("CPI"),
						boshtbl.NewHeader("Features"),
						boshtbl.NewHeader("User"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("director-name"),
							boshtbl.NewValueString("director-uuid"),
							boshtbl.NewValueString("director-version"),
							boshtbl.NewValueString("cpi-info"),
							boshtbl.NewValueStrings([]string{"feature-1: enabled"}),
							boshtbl.NewValueString("best-user"),
						},
					},
					Transpose: true,
				}))
			})
		})

		Context("with minimum director info", func() {
			It("returns a table with the director info", func() {
				info := boshdir.Info{
					Name:    "director-name",
					UUID:    "director-uuid",
					Version: "director-version",
				}
				director.InfoReturns(info, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Name"),
						boshtbl.NewHeader("UUID"),
						boshtbl.NewHeader("Version"),
						boshtbl.NewHeader("User"),
					},
					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("director-name"),
							boshtbl.NewValueString("director-uuid"),
							boshtbl.NewValueString("director-version"),
							boshtbl.NewValueString("(not logged in)"),
						},
					},
					Transpose: true,
				}))
			})
		})

		Context("when director info cannot be fetched", func() {
			It("returns error", func() {
				director.InfoReturns(boshdir.Info{}, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Describe("optional value rendering", func() {
			It("shows CPI information when present", func() {
				info := boshdir.Info{
					CPI: "cpi-info",
				}
				director.InfoReturns(info, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table.Rows[0]).To(ContainElement(boshtbl.NewValueString("cpi-info")))
			})

			It("shows Feature information when present", func() {
				info := boshdir.Info{
					Features: map[string]bool{"feature-1": true},
				}
				director.InfoReturns(info, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table.Rows[0]).To(ContainElement(
					boshtbl.NewValueStrings([]string{"feature-1: enabled"}),
				))
			})
		})

		Context("When details flag is passed", func() {
			BeforeEach(func() {
				opts = EnvironmentOpts{Details: true}
			})

			It("it renders the certificates information", func() {
				output := []boshdir.CertificateExpiryInfo{
					{Path: "foo", Expiry: "2019-11-21T21:43:58Z", DaysLeft: 351},
					{Path: "bar", Expiry: "2018-12-04T21:43:58Z", DaysLeft: 0},
					{Path: "baz", Expiry: "2018-11-21T21:43:58Z", DaysLeft: -5},
				}

				var rows = [][]boshtbl.Value{}
				for _, certificate := range output {
					row := []boshtbl.Value{
						boshtbl.NewValueString(certificate.Path),
						boshtbl.NewValueString(certificate.Expiry),
						boshtbl.NewValueFmt(boshtbl.NewValueInt(certificate.DaysLeft), certificate.DaysLeft <= 30),
					}
					rows = append(rows, row)
				}

				director.CertificateExpiryReturns(output, nil)
				err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(ui.Table).To(Equal(boshtbl.Table{
					Title: color.New(color.Bold).Sprintf(color.YellowString("CERTIFICATE EXPIRY DATE INFORMATION")),
					Header: []boshtbl.Header{
						boshtbl.NewHeader("Certificate"),
						boshtbl.NewHeader("Expiry Date (UTC)"),
						boshtbl.NewHeader("Days Left"),
					},
					Rows:      rows,
					Transpose: false,
				}))
			})
		})
	})
})
