package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
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
		Type: opts.Type,
		Name: opts.Name,
	}

	configs, err := c.director.ListConfigs(opts.Recent, filter)
	if err != nil {
		return err
	}

	var headers []boshtbl.Header
	headers = append(headers, boshtbl.NewHeader("ID"))
	headers = append(headers, boshtbl.NewHeader("Type"))
	headers = append(headers, boshtbl.NewHeader("Name"))
	headers = append(headers, boshtbl.NewHeader("Team"))
	headers = append(headers, boshtbl.NewHeader("Created At"))

	notes := []string{}
	if atLeastOneConfigIsCurrent(configs) {
		notes = append(notes, "(*) Currently active")
	}

	table := boshtbl.Table{
		Content: "configs",
		Header:  headers,
		Notes:   notes,
	}

	if opts.Recent <= 1 {
		table.Notes = append(table.Notes, "Only showing active configs. To see older versions use the --recent=10 option.")
	}

	for _, config := range configs {
		var result []boshtbl.Value
		idString := config.ID

		if config.Current {
			idString += "*"
		}

		result = append(result, boshtbl.NewValueString(idString))
		result = append(result, boshtbl.NewValueString(config.Type))
		result = append(result, boshtbl.NewValueString(config.Name))
		result = append(result, boshtbl.NewValueString(config.Team))
		result = append(result, boshtbl.NewValueString(config.CreatedAt))
		table.Rows = append(table.Rows, result)
	}

	c.ui.PrintTable(table)
	return nil
}

func atLeastOneConfigIsCurrent(configs []boshdir.Config) bool {
	for _, config := range configs {
		if config.Current {
			return true
		}
	}
	return false
}
