package cmd

import (
	"fmt"
	"strings"

	boshrel "github.com/cloudfoundry/bosh-init/release"
	boshrelpkg "github.com/cloudfoundry/bosh-init/release/pkg"
	boshui "github.com/cloudfoundry/bosh-init/ui"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

type ReleaseTables struct {
	Release     boshrel.Release
	ArchivePath string
}

func (t ReleaseTables) Print(ui boshui.UI) {
	summaryTable := boshtbl.Table{
		Rows: [][]boshtbl.Value{
			{
				boshtbl.ValueString{"Name"},
				boshtbl.ValueString{t.Release.Name()},
			},
			{
				boshtbl.ValueString{"Version"},
				boshtbl.ValueString{t.Release.Version()},
			},
			{
				boshtbl.ValueString{"Commit Hash"},
				boshtbl.ValueString{t.Release.CommitHashWithMark("+")},
			},
		},
	}

	if len(t.ArchivePath) > 0 {
		summaryTable.Rows = append(summaryTable.Rows, []boshtbl.Value{
			boshtbl.ValueString{"Archive"},
			boshtbl.ValueString{t.ArchivePath},
		})
	}

	jobsTable := boshtbl.Table{
		Content: "jobs",
		Header:  []string{"Job", "SHA1", "Packages"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, job := range t.Release.Jobs() {
		jobsTable.Rows = append(jobsTable.Rows, []boshtbl.Value{
			boshtbl.ValueString{fmt.Sprintf("%s/%s", job.Name(), job.Fingerprint())},
			boshtbl.ValueString{job.ArchiveSHA1()},
			boshtbl.ValueString{t.sumPkgNames(job.Packages)},
		})
	}

	pkgsTable := boshtbl.Table{
		Content: "packages",
		Header:  []string{"Package", "SHA1", "Dependencies"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, pkg := range t.Release.Packages() {
		pkgsTable.Rows = append(pkgsTable.Rows, []boshtbl.Value{
			boshtbl.ValueString{fmt.Sprintf("%s/%s", pkg.Name(), pkg.Fingerprint())},
			boshtbl.ValueString{pkg.ArchiveSHA1()},
			boshtbl.ValueString{t.sumPkgDependencyNames(pkg.Dependencies)},
		})
	}

	ui.PrintTable(summaryTable)
	ui.PrintTable(jobsTable)
	ui.PrintTable(pkgsTable)
}

func (t ReleaseTables) sumPkgNames(packages []boshrelpkg.Compilable) string {
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.Name())
	}
	return strings.Join(names, ", ")
}

func (t ReleaseTables) sumPkgDependencyNames(packages []*boshrelpkg.Package) string {
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.Name())
	}
	return strings.Join(names, ", ")
}
