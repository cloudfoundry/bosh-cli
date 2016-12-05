package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type VarsCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewVarsCmd(ui boshui.UI, deployment boshdir.Deployment) VarsCmd {
	return VarsCmd{ui: ui, deployment: deployment}
}

func (c VarsCmd) Run() error {

	vars, err := c.deployment.ConfigVars()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Header: []string{"ID", "Name"},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1},
		},
	}

	for _, configVar := range vars {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(configVar.PlaceholderID),
			boshtbl.NewValueString(configVar.PlaceholderName),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
