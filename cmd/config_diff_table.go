package cmd

import (
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ConfigDiffTable struct {
	diff Diff
	opts DiffConfigOpts
	ui   boshui.UI
}

func NewConfigDiffTable(diff Diff, opts DiffConfigOpts, ui boshui.UI) ConfigDiffTable {
	return ConfigDiffTable{diff, opts, ui}
}

func (t ConfigDiffTable) Print() {
	headers := []boshtbl.Header{
		boshtbl.NewHeader("From ID"),
		boshtbl.NewHeader("To ID"),
		boshtbl.NewHeader("Diff"),
	}

	table := boshtbl.Table{
		Content: "",
		Header:  headers,
		Notes:   []string{},

		FillFirstColumn: true,

		Transpose: true,
	}

	result := []boshtbl.Value{
		boshtbl.NewValueString(t.opts.FromID),
		boshtbl.NewValueString(t.opts.ToID),
		boshtbl.NewValueString(t.diff.String()),
	}

	table.Rows = append(table.Rows, result)

	t.ui.PrintTable(table)
}
