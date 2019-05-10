package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type VariablesCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewVariablesCmd(ui boshui.UI, deployment boshdir.Deployment) VariablesCmd {
	return VariablesCmd{ui: ui, deployment: deployment}
}

func (c VariablesCmd) Run(opts VariablesOpts) error {

	variables, err := c.deployment.Variables(opts.Type)
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "variables",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Type"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 1, Asc: true},
		},
	}

	for _, variable := range variables {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(variable.ID),
			boshtbl.NewValueString(variable.Name),
			boshtbl.NewValueString(variable.Type),
		})
	}

	c.ui.PrintTable(table)

	return nil
}

func (c VariablesCmd) getCertificate() error {

	variableCerts, err := c.deployment.VariableCerts()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
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
	}

	for _, varCert := range variableCerts {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(varCert.ID),
			boshtbl.NewValueString(varCert.Name),
			boshtbl.NewValueString(varCert.ExpiryDate),
			boshtbl.NewValueInt(varCert.DaysLeft),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
