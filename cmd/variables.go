package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type VariablesCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewVariablesCmd(ui boshui.UI, deployment boshdir.Deployment) VariablesCmd {
	return VariablesCmd{ui: ui, deployment: deployment}
}

func (c VariablesCmd) Run(opts VariablesOpts) error {
	variables, err := c.deployment.Variables()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "variables",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("Name"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 1, Asc: true},
		},
	}

	for _, variable := range variables {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(variable.ID),
			boshtbl.NewValueString(variable.Name),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
