package cmd

import (
	"fmt"
	"strings"

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
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Version"),
			boshtbl.NewHeader("Commit Hash"),
		},
		Rows: [][]boshtbl.Value{
			{
				boshtbl.NewValueString(t.Release.Name()),
				boshtbl.NewValueString(t.Release.Version()),
				boshtbl.NewValueString(t.Release.CommitHashWithMark("+")),
			},
		},
		Transpose: true,
	}

	if len(t.ArchivePath) > 0 {
		summaryTable = summaryTable.AddColumn("Archive", []boshtbl.Value{
			boshtbl.NewValueString(t.ArchivePath),
		})
	}

	jobsTable := boshtbl.Table{
		Content: "jobs",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Job"),
			boshtbl.NewHeader("Digest"),
			boshtbl.NewHeader("Packages"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, job := range t.Release.Jobs() {
		jobsTable.Rows = append(jobsTable.Rows, []boshtbl.Value{
			boshtbl.NewValueString(fmt.Sprintf("%s/%s", job.Name(), job.Fingerprint())),
			boshtbl.NewValueString(job.ArchiveDigest()),
			boshtbl.NewValueStrings(t.sumPkgNames(job.Packages)),
		})
	}

	compiledPackages := t.Release.CompiledPackages()

	pkgsTable := boshtbl.Table{
		Content: "packages",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Package"),
			boshtbl.NewHeader("Digest"),
			boshtbl.NewHeader("Dependencies"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	if len(compiledPackages) > 0 {
		pkgsTable.Header = append(
			pkgsTable.Header,
			boshtbl.NewHeader("OS"),
			boshtbl.NewHeader("OS Version"),
		)
	}

	for _, pkg := range t.Release.Packages() {
		row := []boshtbl.Value{
			boshtbl.NewValueString(fmt.Sprintf("%s/%s", pkg.Name(), pkg.Fingerprint())),
			boshtbl.NewValueString(pkg.ArchiveDigest()),
			boshtbl.NewValueStrings(t.sumPkgNames(pkg.Deps())),
		}
		if len(compiledPackages) > 0 {
			row = append(
				row,
				boshtbl.NewValueString(""),
				boshtbl.NewValueString(""),
			)
		}
		pkgsTable.Rows = append(pkgsTable.Rows, row)
	}

	for _, pkg := range compiledPackages {
		osParts := strings.Split(pkg.OSVersionSlug(), "/")
		osName := osParts[0]
		osVersion := ""
		if len(osParts) > 1 {
			osVersion = osParts[1]
		}

		pkgsTable.Rows = append(pkgsTable.Rows, []boshtbl.Value{
			boshtbl.NewValueString(fmt.Sprintf("%s/%s", pkg.Name(), pkg.Fingerprint())),
			boshtbl.NewValueString(pkg.ArchiveDigest()),
			boshtbl.NewValueStrings(t.sumPkgNames(pkg.Deps())),
			boshtbl.NewValueString(osName),
			boshtbl.NewValueString(osVersion),
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
