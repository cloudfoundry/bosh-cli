package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	"github.com/fatih/color"
)

type CertificateInfoTable struct {
	Certificates []boshdir.CertificateExpiryInfo
	UI           boshui.UI
}

func (t CertificateInfoTable) Print() {

	var rows [][]boshtbl.Value

	for _, certificate := range t.Certificates {
		status := ""

		if certificate.DaysLeft > 30 {
			status = color.GreenString("valid")
		} else if certificate.DaysLeft >= 0 {
			status = color.YellowString("expiring")
		} else {
			status = color.RedString("expired")
		}

		row := []boshtbl.Value{
			boshtbl.NewValueString(certificate.Path),
			boshtbl.NewValueString(certificate.Expiry),
			boshtbl.NewValueInt(certificate.DaysLeft),
			boshtbl.NewValueString(status),
		}

		rows = append(rows, row)
	}

	table := boshtbl.Table{
		Title: color.New(color.Bold).Sprintf(color.YellowString("CERTIFICATE EXPIRY DATE INFORMATION")),
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Certificate"),
			boshtbl.NewHeader("Expiry Date (UTC)"),
			boshtbl.NewHeader("Days Left"),
			boshtbl.NewHeader("Status"),
		},
		Rows:      rows,
		Transpose: false,
	}

	t.UI.PrintTable(table)
}
