package cmd

import (
	"fmt"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type InspectReleaseCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewInspectReleaseCmd(ui boshui.UI, director boshdir.Director) InspectReleaseCmd {
	return InspectReleaseCmd{ui: ui, director: director}
}

func (c InspectReleaseCmd) Run(opts InspectReleaseOpts) error {
	release, err := c.director.FindRelease(opts.Args.Slug)
	if err != nil {
		return err
	}

	jobsTable := boshtbl.Table{
		Content: "jobs",
		Header:  []string{"Job", "Blobstore ID", "SHA1"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	jobs, err := release.Jobs()
	if err != nil {
		return err
	}

	for _, j := range jobs {
		jobsTable.Rows = append(jobsTable.Rows, []boshtbl.Value{
			boshtbl.ValueString{fmt.Sprintf("%s/%s", j.Name, j.Fingerprint)},
			boshtbl.ValueString{j.BlobstoreID},
			boshtbl.ValueString{j.SHA1},
		})
	}

	pkgsTable := boshtbl.Table{
		Content: "packages",
		Header:  []string{"Package", "Compiled for", "Blobstore ID", "SHA1"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	pkgs, err := release.Packages()
	if err != nil {
		return err
	}

	for _, p := range pkgs {
		section := boshtbl.Section{
			FirstColumn: boshtbl.ValueString{fmt.Sprintf("%s/%s", p.Name, p.Fingerprint)},

			Rows: [][]boshtbl.Value{
				{
					boshtbl.ValueString{""},
					boshtbl.ValueString{"(source)"},
					boshtbl.ValueString{p.BlobstoreID},
					boshtbl.ValueString{p.SHA1},
				},
			},
		}

		for _, cp := range p.CompiledPackages {
			section.Rows = append(section.Rows, []boshtbl.Value{
				boshtbl.ValueString{""},
				boshtbl.ValueString{cp.StemcellSlug.String()},
				boshtbl.ValueString{cp.BlobstoreID},
				boshtbl.ValueString{cp.SHA1},
			})
		}

		pkgsTable.Sections = append(pkgsTable.Sections, section)
	}

	c.ui.PrintTable(jobsTable)
	c.ui.PrintTable(pkgsTable)

	return nil
}
