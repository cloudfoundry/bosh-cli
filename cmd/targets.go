package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type TargetsCmd struct {
	config cmdconf.Config
	ui     boshui.UI
}

func NewTargetsCmd(config cmdconf.Config, ui boshui.UI) TargetsCmd {
	return TargetsCmd{config: config, ui: ui}
}

func (c TargetsCmd) Run() error {
	targets := c.config.Targets()

	table := boshtbl.Table{
		Content: "targets",
		Header:  []string{"URL", "Alias"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, t := range targets {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(t.URL),
			boshtbl.NewValueString(t.Alias),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
