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
		row := []boshtbl.Value{
			boshtbl.NewValueString(certificate.Path),
			boshtbl.NewValueString(certificate.Expiry),
			boshtbl.NewValueFmt(boshtbl.NewValueInt(certificate.DaysLeft), certificate.DaysLeft <= 30),
		}
		rows = append(rows, row)
	}

	table := boshtbl.Table{
		Title: color.New(color.Bold).Sprintf(color.YellowString("CERTIFICATE EXPIRY DATE INFORMATION")),
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Certificate"),
			boshtbl.NewHeader("Expiry Date (UTC)"),
			boshtbl.NewHeader("Days Left"),
		},
		Rows:      rows,
		Transpose: false,
	}

	t.UI.PrintTable(table)
}
