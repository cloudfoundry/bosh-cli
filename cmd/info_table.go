package cmd

import (
	"fmt"
	"sort"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type InfoTable struct {
	Info boshdir.Info
	UI   boshui.UI
}

func (t InfoTable) Print() {
	table := boshtbl.Table{
		Rows: [][]boshtbl.Value{
			{
				boshtbl.ValueString{"Name"},
				boshtbl.ValueString{t.Info.Name},
			},
			{
				boshtbl.ValueString{"UUID"},
				boshtbl.ValueString{t.Info.UUID},
			},
			{
				boshtbl.ValueString{"Version"},
				boshtbl.ValueString{t.Info.Version},
			},
		},
	}

	if len(t.Info.CPI) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.ValueString{"CPI"},
			boshtbl.ValueString{t.Info.CPI},
		})
	}

	if len(t.Info.Features) > 0 {
		desc := []string{}

		enabledText := map[bool]string{
			true:  "enabled",
			false: "disabled",
		}

		for name, enabled := range t.Info.Features {
			desc = append(desc, fmt.Sprintf("%s: %s", name, enabledText[enabled]))
		}

		sort.Sort(InfoFeatureSorting(desc))

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.ValueString{"Features"},
			boshtbl.ValueStrings{desc},
		})
	}

	if len(t.Info.User) > 0 {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.ValueString{"User"},
			boshtbl.ValueString{t.Info.User},
		})
	} else {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.ValueString{"User"},
			boshtbl.ValueString{"(not logged in)"},
		})
	}

	t.UI.PrintTable(table)
}

type InfoFeatureSorting []string

func (s InfoFeatureSorting) Len() int           { return len(s) }
func (s InfoFeatureSorting) Less(i, j int) bool { return s[i] < s[j] }
func (s InfoFeatureSorting) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
