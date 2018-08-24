package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
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
		ui: ui,
	}
}

func (c InspectStemcellTarballCmd) Run(opts InspectStemcellTarballOpts) error {

	//archive := c.stemcellArchiveFactory(path)
	//
	//name, version, err := archive.Info()
	//if err != nil {
	//	return bosherr.WrapErrorf(err, "Retrieving stemcell info")
	//}

	metadataTable := boshtbl.Table{
		Content: "stemcell-metadata",
		Header: []boshtbl.Header{
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("OS"),
			boshtbl.NewHeader("Version"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
		Rows: [][]boshtbl.Value{
			{
				boshtbl.NewValueString("example-name"),
				boshtbl.NewValueString("example-os"),
				boshtbl.NewValueString("example.version"),
			},
		},
	}

	c.ui.PrintTable(metadataTable)
	return nil
}
