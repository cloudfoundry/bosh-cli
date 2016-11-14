package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type InstancesCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewInstancesCmd(ui boshui.UI, deployment boshdir.Deployment) InstancesCmd {
	return InstancesCmd{ui: ui, deployment: deployment}
}

func (c InstancesCmd) Run(opts InstancesOpts) error {
	instanceInfos, err := c.deployment.InstanceInfos()
	if err != nil {
		return err
	}

	instTable := InstanceTable{
		Processes: opts.Processes,
		Details:   opts.Details,
		DNS:       opts.DNS,
		Vitals:    opts.Vitals,
	}

	table := boshtbl.Table{
		Content: "instances",

		HeaderVals: instTable.AsValues(instTable.Header()),

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true}, // sort by process so that VM row is first
		},

		Notes: []string{""},
	}

	for _, info := range instanceInfos {
		if opts.Failing && info.IsRunning() {
			continue
		}

		row := instTable.AsValues(instTable.ForVMInfo(info))

		section := boshtbl.Section{
			FirstColumn: row[0],
			Rows:        [][]boshtbl.Value{row},
		}

		if opts.Processes {
			for _, p := range info.Processes {
				if opts.Failing && p.IsRunning() {
					continue
				}

				row := instTable.AsValues(instTable.ForProcess(p))

				section.Rows = append(section.Rows, row)
			}
		}

		table.Sections = append(table.Sections, section)
	}

	c.ui.PrintTable(table)

	return nil
}
