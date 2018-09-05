package cmd

import (
	"fmt"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type DeploymentTablePrinter struct {
	Name      string
	Releases  []boshdir.Release
	Stemcells []boshdir.Stemcell
	Teams     []string
	Configs   boshdir.DeploymentConfigs
	UI        boshui.UI
}

func (t DeploymentTablePrinter) Print() error {
	headers := []boshtbl.Header{
		boshtbl.NewHeader("Name"),
		boshtbl.NewHeader("Release(s)"),
		boshtbl.NewHeader("Stemcell(s)"),
		boshtbl.NewHeader("Config(s)"),
		boshtbl.NewHeader("Team(s)"),
	}

	rows := []boshtbl.Value{
		boshtbl.NewValueString(t.Name),
		boshtbl.NewValueStrings(formatReleases(t.Releases)),
		boshtbl.NewValueStrings(formatStemcells(t.Stemcells)),
		boshtbl.NewValueStrings(formatConfigs(t.Configs)),
		boshtbl.NewValueStrings(t.Teams),
	}

	table := boshtbl.Table{
		Content: "deployments",

		Header: headers,

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	table.Rows = append(table.Rows, rows)

	t.UI.PrintTable(table)

	return nil
}

func formatReleases(rels []boshdir.Release) []string {
	var names []string
	for _, r := range rels {
		names = append(names, r.Name()+"/"+r.Version().String())
	}
	return names
}

func formatStemcells(stemcells []boshdir.Stemcell) []string {
	var names []string
	for _, s := range stemcells {
		names = append(names, s.Name()+"/"+s.Version().String())
	}
	return names
}

func formatConfigs(configs boshdir.DeploymentConfigs) []string {
	var names []string
	for _, c := range configs.GetConfigs() {
		names = append(names, fmt.Sprintf("%d %s/%s", c.Id, c.Type, c.Name))
	}
	return names
}
