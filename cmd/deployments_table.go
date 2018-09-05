package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"
)

type DeploymentsTable struct {
	Deployments []boshdir.DeploymentResp
	UI          boshui.UI
}

func (t DeploymentsTable) Print() error {
	table := boshtbl.Table{
		Content: "deployments",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Release(s)"),
			boshtbl.NewHeader("Stemcell(s)"),
			boshtbl.NewHeader("Team(s)"),
		},

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
		},
	}

	for _, d := range t.Deployments {
		releases, err := takeReleases(d.Releases)
		if err != nil {
			return err
		}
		stemcells, err := takeStemcells(d.Stemcells)
		if err != nil {
			return err
		}

		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueString(d.Name),
			boshtbl.NewValueStrings(releases),
			boshtbl.NewValueStrings(stemcells),
			boshtbl.NewValueStrings(d.Teams),
		})
	}

	t.UI.PrintTable(table)

	return nil
}

func takeReleases(rels []boshdir.DeploymentReleaseResp) ([]string, error) {
	var names []string
	for _, r := range rels {
		parsedVersion, err := semver.NewVersionFromString(r.Version)
		if err != nil {
			return nil, bosherr.WrapErrorf(
				err, "Parsing version for release '%s/%s'", r.Name, r.Version)
		}
		names = append(names, r.Name+"/"+parsedVersion.String())
	}
	return names, nil
}

func takeStemcells(stemcells []boshdir.DeploymentStemcellResp) ([]string, error) {
	var names []string
	for _, s := range stemcells {
		parsedVersion, err := semver.NewVersionFromString(s.Version)
		if err != nil {
			return nil, bosherr.WrapErrorf(
				err, "Parsing version for stemcell '%s/%s'", s.Name, s.Version)
		}
		names = append(names, s.Name+"/"+parsedVersion.String())
	}
	return names, nil
}
