package cmd_test

import (
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("CertificateInfoTable", func() {
	var (
		ui *fakeui.FakeUI
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
	})

	Describe("Print", func() {
		It("shows information about the director's certificates expiry", func() {
			certsInfo := []boshdir.CertificateExpiryInfo{
				{Path: "foo", Expiry: "2019-11-21T21:43:57Z", DaysLeft: 353},
				{Path: "bar", Expiry: "2020-11-21T21:43:57Z", DaysLeft: 0},
				{Path: "baz", Expiry: "2018-10-21T21:43:57Z", DaysLeft: -10},
			}

			CertificateInfoTable{Certificates: certsInfo, UI: ui}.Print()

			Expect(ui.Table.Title).To(Equal(color.New(color.Bold).Sprintf(color.YellowString("CERTIFICATE EXPIRY DATE INFORMATION"))))
			Expect(ui.Table.Header).To(Equal([]boshtbl.Header{
				boshtbl.NewHeader("Certificate"),
				boshtbl.NewHeader("Expiry Date (UTC)"),
				boshtbl.NewHeader("Days Left"),
			}))
			Expect(ui.Table.Rows).To(HaveLen(3))

			for i, certificate := range certsInfo {
				Expect(ui.Table.Rows[i]).To(Equal([]boshtbl.Value{
					boshtbl.NewValueString(certificate.Path),
					boshtbl.NewValueString(certificate.Expiry),
					boshtbl.NewValueFmt(boshtbl.NewValueInt(certificate.DaysLeft), certificate.DaysLeft <= 30),
				}))
			}
		})
	})
})
