package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

type InspectStemcellTarballCmd struct {
	stemcellArchiveFactory func(string) boshdir.StemcellArchive
	ui                     biui.UI
}

func NewInspectStemcellTarballCmd(
	stemcellArchiveFactory func(string) boshdir.StemcellArchive,
	ui biui.UI,
) InspectStemcellTarballCmd {
	return InspectStemcellTarballCmd{
		stemcellArchiveFactory: stemcellArchiveFactory,
		ui:                     ui,
	}
}

func (c InspectStemcellTarballCmd) Run(opts InspectStemcellTarballOpts) error {
	archive := c.stemcellArchiveFactory(opts.Args.PathToStemcell)
	metadata, err := archive.Info()
	if err != nil {
		return err
	}

	infrastructure := metadata.CloudProperties["infrastructure"]
	if infrastructure == nil {
		infrastructure = "unknown"
	}

	hypervisor := metadata.CloudProperties["hypervisor"]
	if hypervisor == nil {
		hypervisor = "-"
	}

	metadataTable := boshtbl.Table{
		Content: "stemcell-metadata",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("OS"),
			boshtbl.NewHeader("Version"),
			boshtbl.NewHeader("Infrastructure"),
			boshtbl.NewHeader("Hypervisor"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
		Rows: [][]boshtbl.Value{
			{
				boshtbl.NewValueString(metadata.Name),
				boshtbl.NewValueString(metadata.OS),
				boshtbl.NewValueString(metadata.Version),
				boshtbl.NewValueString(infrastructure.(string)),
				boshtbl.NewValueString(hypervisor.(string)),
			},
		},
	}

	c.ui.PrintTable(metadataTable)
	return nil
}
