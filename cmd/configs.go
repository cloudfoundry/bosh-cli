package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

type ConfigsCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewConfigsCmd(ui boshui.UI, director boshdir.Director) ConfigsCmd {
	return ConfigsCmd{ui: ui, director: director}
}

func (c ConfigsCmd) Run(opts ConfigsOpts) error {
	filter := boshdir.ConfigsFilter{
		Type:            opts.Type,
		Name:            opts.Name,
		IncludeOutdated: opts.IncludeOutdated,
	}

	configs, err := c.director.ListConfigs(filter)
	if err != nil {
		return err
	}

	var headers []boshtbl.Header
	if filter.IncludeOutdated {
		headers = append(headers, boshtbl.NewHeader("ID"))
	}
	headers = append(headers, boshtbl.NewHeader("Type"))
	headers = append(headers, boshtbl.NewHeader("Name"))

	table := boshtbl.Table{
		Content: "configs",
		Header:  headers,
	}

	for _, config := range configs {
		var result []boshtbl.Value
		if filter.IncludeOutdated {
			result = append(result, boshtbl.NewValueString(config.Id))
		}
		result = append(result, boshtbl.NewValueString(config.Type))
		result = append(result, boshtbl.NewValueString(config.Name))
		table.Rows = append(table.Rows, result)
	}

	c.ui.PrintTable(table)
	return nil
}
