package cmd

import (
	"github.com/cloudfoundry/bosh-cli/director"
	"github.com/cloudfoundry/bosh-cli/ui"
	"github.com/cloudfoundry/bosh-cli/ui/table"
)

type OrphanedVMsCmd struct {
	ui       ui.UI
	director director.Director
}

func NewOrphanedVMsCmd(ui ui.UI, director director.Director) OrphanedVMsCmd {
	return OrphanedVMsCmd{ui: ui, director: director}
}

func (c OrphanedVMsCmd) Run() error {
	orphanedVMs, err := c.director.OrphanedVMs()
	if err != nil {
		return err
	}

	tbl := table.Table{
		Content: "orphaned_vms",
		Header: []table.Header{
			table.NewHeader("VM CID"),
			table.NewHeader("Deployment"),
			table.NewHeader("Instance"),
			table.NewHeader("AZ"),
			table.NewHeader("IPs"),
			table.NewHeader("Orphaned At"),
		},
		SortBy: []table.ColumnSort{{Column: 5}},
	}

	for _, vm := range orphanedVMs {
		tbl.Rows = append(tbl.Rows, []table.Value{
			table.NewValueString(vm.CID),
			table.NewValueString(vm.DeploymentName),
			table.NewValueString(vm.InstanceName),
			table.NewValueString(vm.AZName),
			table.NewValueStrings(vm.IPAddresses),
			table.NewValueTime(vm.OrphanedAt),
		})
	}

	c.ui.PrintTable(tbl)

	return nil
}
