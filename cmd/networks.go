package cmd

import (
	"errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type NetworksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewNetworksCmd(ui boshui.UI, director boshdir.Director) NetworksCmd {
	return NetworksCmd{ui: ui, director: director}
}

func (c NetworksCmd) Run(opts NetworksOpts) error {
	if !opts.Orphaned {
		return errors.New("Only --orphaned is supported")
	}

	networks, err := c.director.OrphanNetworks()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "networks",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Type"),
			boshtbl.NewHeader("Created At"),
			boshtbl.NewHeader("Orphaned At"),
		},
		SortBy: []boshtbl.ColumnSort{{
			Column: 0,
			Asc:    true,
		}},
	}

	for _, n := range networks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(n.Name()),
			boshtbl.NewValueString(n.Type()),
			boshtbl.NewValueTime(n.CreatedAt()),
			boshtbl.NewValueTime(n.OrphanedAt()),
		})
	}

	c.ui.PrintTable(table)

	return nil
}
