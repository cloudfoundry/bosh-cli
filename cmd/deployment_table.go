package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type DeploymentTable struct {
	Deployments []boshdir.Deployment
	UI          boshui.UI
}

func (t DeploymentTable) Print() error {
	table := boshtbl.Table{
		Content: "deployments",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Release(s)"),
			boshtbl.NewHeader("Stemcell(s)"),
			boshtbl.NewHeader("Team(s)"),
			boshtbl.NewHeader("Cloud Config"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	for _, d := range t.Deployments {
		releases, err := d.Releases()
		if err != nil {
			return err
		}

		stemcells, err := d.Stemcells()
		if err != nil {
			return err
		}

		teams, err := d.Teams()
		if err != nil {
			return err
		}

		cloudConfig, err := d.CloudConfig()
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(d.Name()),
			boshtbl.NewValueStrings(t.takeReleases(releases)),
			boshtbl.NewValueStrings(t.takeStemcells(stemcells)),
			boshtbl.NewValueStrings(teams),
			boshtbl.NewValueString(cloudConfig),
		})
	}

	t.UI.PrintTable(table)

	return nil
}

func (t DeploymentTable) takeReleases(rels []boshdir.Release) []string {
	var names []string
	for _, r := range rels {
		names = append(names, r.Name()+"/"+r.Version().String())
	}
	return names
}

func (t DeploymentTable) takeStemcells(stemcells []boshdir.Stemcell) []string {
	var names []string
	for _, s := range stemcells {
		names = append(names, s.Name()+"/"+s.Version().String())
	}
	return names
}
