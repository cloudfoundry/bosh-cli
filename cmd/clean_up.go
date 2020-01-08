package cmd

import (
	"strconv"

	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type CleanUpCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCleanUpCmd(ui boshui.UI, director boshdir.Director) CleanUpCmd {
	return CleanUpCmd{ui: ui, director: director}
}

func (c CleanUpCmd) Run(opts CleanUpOpts) error {

	if opts.DryRun == false {
		err := c.ui.AskForConfirmation()
		if err != nil {
			return err
		}
	}

	resp, err := c.director.CleanUp(opts.All, opts.DryRun, opts.KeepOrphanedDisks)
	if err != nil {
		return err
	}

	if opts.DryRun {
		c.PrintCleanUpTable(resp)
	}
	return nil
}

func (c CleanUpCmd) PrintCleanUpTable(resp boshdir.CleanUp) {
	titles := []string{
		"Unused Releases",
		"Unused Stemcells",
		"Unused Compiled Packages",
		"Exported Releases",
		"Stale DNS Record Blobs",
		"Orphaned Disks",
		"Orphaned VMs",
	}

	releaseHeaders := boshtbl.NewHeadersFromStrings([]string{"Name", "Version"})
	stemcellHeaders := boshtbl.NewHeadersFromStrings([]string{"Name", "Version"})
	compiledPackageHeaders := boshtbl.NewHeadersFromStrings([]string{"Name", "Stemcell OS", "Stemcell Version"})
	exportedReleasesHeaders := []boshtbl.Header{boshtbl.NewHeader("Blob ID")}
	dnsRecordBlobsHeaders := exportedReleasesHeaders
	orphanedDisksHeaders := boshtbl.NewHeadersFromStrings([]string{"Disk CID", "Deployment", "Instance", "Size (mb)"})
	orphanedVmsHeaders := boshtbl.NewHeadersFromStrings([]string{"VM CID", "Deployment", "Instance"})

	headers := [][]boshtbl.Header{
		releaseHeaders,
		stemcellHeaders,
		compiledPackageHeaders,
		exportedReleasesHeaders,
		dnsRecordBlobsHeaders,
		orphanedDisksHeaders,
		orphanedVmsHeaders,
	}

	rows := [][][]boshtbl.Value{
		releaseRows(resp.Releases),
		stemcellRows(resp.Stemcells),
		compiledPackageRows(resp.CompiledPackages),
		stringBasedRows(resp.ExportedReleases),
		stringBasedRows(resp.DNSBlobs),
		orphanedDiskRows(resp.OrphanedDisks),
		orphanedVmRows(resp.OrphanedVMs),
	}

	for i, _ := range titles {
		table := boshtbl.Table{
			Title:  titles[i],
			Header: headers[i],
			SortBy: []boshtbl.ColumnSort{
				{Column: 0, Asc: true},
			},
		}
		for _, r := range rows[i] {
			table.Rows = append(table.Rows, r)
		}
		c.ui.PrintTable(table)
	}
}

func releaseRows(releases []boshdir.CleanableRelease) [][]boshtbl.Value {
	var rows [][]boshtbl.Value
	for _, release := range releases {
		for _, val := range release.Versions {
			row := []boshtbl.Value{
				boshtbl.NewValueString(release.Name),
				boshtbl.NewValueString(val),
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func stemcellRows(stemcells []boshdir.Stemcell) [][]boshtbl.Value {
	var rows [][]boshtbl.Value
	for _, stemcell := range stemcells {
		row := []boshtbl.Value{
			boshtbl.NewValueString(stemcell.Name()),
			boshtbl.NewValueString(stemcell.Version().String()),
		}
		rows = append(rows, row)
	}
	return rows
}

func compiledPackageRows(packages []boshdir.CleanableCompiledPackage) [][]boshtbl.Value {
	var rows [][]boshtbl.Value
	for _, pkg := range packages {
		row := []boshtbl.Value{
			boshtbl.NewValueString(pkg.Name),
			boshtbl.NewValueString(pkg.StemcellOs),
			boshtbl.NewValueString(pkg.StemcellVersion),
		}
		rows = append(rows, row)
	}
	return rows
}

func orphanedVmRows(items []boshdir.OrphanedVM) [][]boshtbl.Value {
	var rows [][]boshtbl.Value
	for _, item := range items {
		row := []boshtbl.Value{
			boshtbl.NewValueString(item.CID),
			boshtbl.NewValueString(item.DeploymentName),
			boshtbl.NewValueString(item.InstanceName),
		}
		rows = append(rows, row)
	}
	return rows
}

func orphanedDiskRows(items []boshdir.OrphanDiskResp) [][]boshtbl.Value {
	var rows [][]boshtbl.Value
	for _, item := range items {
		row := []boshtbl.Value{
			boshtbl.NewValueString(item.CID),
			boshtbl.NewValueString(item.DeploymentName),
			boshtbl.NewValueString(item.InstanceName),
			boshtbl.NewValueString(strconv.FormatUint(item.Size, 10)),
		}
		rows = append(rows, row)
	}
	return rows
}
func stringBasedRows(items []string) [][]boshtbl.Value {
	var rows [][]boshtbl.Value
	for _, item := range items {
		row := []boshtbl.Value{
			boshtbl.NewValueString(item),
		}
		rows = append(rows, row)
	}
	return rows
}
