package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	"github.com/cloudfoundry/bosh-cli/v7/director"
	"github.com/cloudfoundry/bosh-cli/v7/ui"
	"github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type OrphanedVMsCmd struct {
	ui       ui.UI
	director director.Director
}

func NewOrphanedVMsCmd(ui ui.UI, director director.Director) OrphanedVMsCmd {
	return OrphanedVMsCmd{ui: ui, director: director}
}

func (c OrphanedVMsCmd) Run(opts OrphanedVMsOpts) error {
	orphanedVMs, err := c.director.OrphanedVMs()
	if err != nil {
		return err
	}

	printOrphanedVmTable(c.ui, orphanedVMs, opts)

	return nil
}

func formatIPAddress(orphanedVM director.OrphanedVM, opts OrphanedVMsOpts) []string {
	if opts.Cidr {
		return orphanedVM.IPAddressesCidr
	}
	return orphanedVM.IPAddresses
}

func printOrphanedVmTable(ui ui.UI, orphanedVMs []director.OrphanedVM, opts OrphanedVMsOpts) {
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
			table.NewValueStrings(formatIPAddress(vm, opts)),
			table.NewValueTime(vm.OrphanedAt),
		})
	}

	ui.PrintTable(tbl)
}
