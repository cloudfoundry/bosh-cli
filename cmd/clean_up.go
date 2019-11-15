package cmd

import (
	"fmt"

	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type CleanUpCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCleanUpCmd(ui boshui.UI, director boshdir.Director) CleanUpCmd {
	return CleanUpCmd{ui: ui, director: director}
}

func (c CleanUpCmd) Run(opts CleanUpOpts) error {

	if opts.DryRun == false {
		err := c.ui.AskForConfirmation()
		if err != nil {
			return err
		}
	}

	resp, err := c.director.CleanUp(opts.All, opts.DryRun)
	if err != nil {
		return err
	}

	if opts.DryRun {
		c.PrintCleanUpTable(resp)
	}
	return nil
}

func (c CleanUpCmd) PrintCleanUpTable(resp boshdir.CleanUp) {
	headers := []boshtbl.Header{
		boshtbl.NewHeader("Releases"),
		boshtbl.NewHeader("Stemcells"),
		boshtbl.NewHeader("Compiled Packages"),
		boshtbl.NewHeader("Orphaned Disks"),
		boshtbl.NewHeader("Orphaned VMs"),
		boshtbl.NewHeader("Exported Releases"),
		boshtbl.NewHeader("DNS Blobs"),
	}
	table := boshtbl.Table{
		Title:   fmt.Sprintf("Cleanup Artifacts"),
		Content: "cleanup",

		Header: headers,

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},
	}
	rows := []boshtbl.Value{
		boshtbl.NewValueStrings(resp.Releases),
		boshtbl.NewValueStrings(resp.Stemcells),
		boshtbl.NewValueStrings(resp.CompiledPackages),
		boshtbl.NewValueStrings(resp.OrphanedDisks),
		boshtbl.NewValueStrings(resp.OrphanedVMs),
		boshtbl.NewValueStrings(resp.ExportedReleases),
		boshtbl.NewValueStrings(resp.DNSBlobs),
	}
	table.Rows = append(table.Rows, rows)

	c.ui.PrintTable(table)
}
