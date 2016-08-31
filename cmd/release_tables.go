package cmd

import (
	"fmt"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshrelpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ReleaseTables struct {
	Release     boshrel.Release
	ArchivePath string
}

func (t ReleaseTables) Print(ui boshui.UI) {
	summaryTable := boshtbl.Table{
		Rows: [][]boshtbl.Value{
			{
				boshtbl.NewValueString("Name"),
				boshtbl.NewValueString(t.Release.Name()),
			},
			{
				boshtbl.NewValueString("Version"),
				boshtbl.NewValueString(t.Release.Version()),
			},
			{
				boshtbl.NewValueString("Commit Hash"),
				boshtbl.NewValueString(t.Release.CommitHashWithMark("+")),
			},
		},
	}

	if len(t.ArchivePath) > 0 {
		summaryTable.Rows = append(summaryTable.Rows, []boshtbl.Value{
			boshtbl.NewValueString("Archive"),
			boshtbl.NewValueString(t.ArchivePath),
		})
	}

	jobsTable := boshtbl.Table{
		Content: "jobs",
		Header:  []string{"Job", "SHA1", "Packages"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, job := range t.Release.Jobs() {
		jobsTable.Rows = append(jobsTable.Rows, []boshtbl.Value{
			boshtbl.NewValueString(fmt.Sprintf("%s/%s", job.Name(), job.Fingerprint())),
			boshtbl.NewValueString(job.ArchiveSHA1()),
			boshtbl.NewValueStrings(t.sumPkgNames(job.Packages)),
		})
	}

	pkgsTable := boshtbl.Table{
		Content: "packages",
		Header:  []string{"Package", "SHA1", "Dependencies"},
		SortBy:  []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, pkg := range t.Release.Packages() {
		pkgsTable.Rows = append(pkgsTable.Rows, []boshtbl.Value{
			boshtbl.NewValueString(fmt.Sprintf("%s/%s", pkg.Name(), pkg.Fingerprint())),
			boshtbl.NewValueString(pkg.ArchiveSHA1()),
			boshtbl.NewValueStrings(t.sumPkgDependencyNames(pkg.Dependencies)),
		})
	}

	ui.PrintTable(summaryTable)
	ui.PrintTable(jobsTable)
	ui.PrintTable(pkgsTable)
}

func (t ReleaseTables) sumPkgNames(packages []boshrelpkg.Compilable) []string {
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.Name())
	}
	return names
}

func (t ReleaseTables) sumPkgDependencyNames(packages []*boshrelpkg.Package) []string {
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.Name())
	}
	return names
}
