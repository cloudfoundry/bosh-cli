package cmd

import (
	"errors"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type DisksCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDisksCmd(ui boshui.UI, director boshdir.Director) DisksCmd {
	return DisksCmd{ui: ui, director: director}
}

func (c DisksCmd) Run(opts DisksOpts) error {
	if !opts.Orphaned {
		return errors.New("Only --orphaned is supported")
	}

	disks, err := c.director.OrphanedDisks()
	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "disks",
		Header:  []string{"Disk CID", "Size", "Deployment", "Instance", "AZ", "Orphaned At"},
		SortBy:  []boshtbl.ColumnSort{{Column: 5}},
	}

	for _, d := range disks {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.ValueString{d.CID()},
			boshtbl.ValueBytes{d.Size()},
			boshtbl.ValueString{d.Deployment().Name()},
			boshtbl.ValueString{d.InstanceName()},
			boshtbl.ValueString{d.AZName()},
			boshtbl.ValueTime{d.OrphanedAt()},
		})
	}

	c.ui.PrintTable(table)

	return nil
}
