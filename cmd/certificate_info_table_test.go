package cmd_test

import (
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("CertificateInfoTable", func() {
	var (
		testUI *testui.Ui
	)

	BeforeEach(func() {
		testUI = &testui.Ui{}
	})

	Describe("Print", func() {
		It("shows information about the director's certificates expiry", func() {
			certsInfo := []boshdir.CertificateExpiryInfo{
				{Path: "foo", Expiry: "2019-11-21T21:43:57Z", DaysLeft: 353},
				{Path: "bar", Expiry: "2020-11-21T21:43:57Z", DaysLeft: 0},
				{Path: "baz", Expiry: "2018-10-21T21:43:57Z", DaysLeft: -10},
			}

			cmd.CertificateInfoTable{Certificates: certsInfo, UI: testUI}.Print()

			Expect(testUI.Table.Title).To(Equal(color.New(color.Bold, color.FgYellow).Sprint("CERTIFICATE EXPIRY DATE INFORMATION")))
			Expect(testUI.Table.Header).To(Equal([]boshtbl.Header{
				boshtbl.NewHeader("Certificate"),
				boshtbl.NewHeader("Expiry Date (UTC)"),
				boshtbl.NewHeader("Days Left"),
			}))
			Expect(testUI.Table.Rows).To(HaveLen(3))

			for i, certificate := range certsInfo {
				Expect(testUI.Table.Rows[i]).To(Equal([]boshtbl.Value{
					boshtbl.NewValueString(certificate.Path),
					boshtbl.NewValueString(certificate.Expiry),
					boshtbl.NewValueFmt(boshtbl.NewValueInt(certificate.DaysLeft), certificate.DaysLeft <= 30),
				}))
			}
		})
	})
})
