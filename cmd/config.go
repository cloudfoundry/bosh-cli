package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewConfigCmd(ui boshui.UI, director boshdir.Director) ConfigCmd {
	return ConfigCmd{ui: ui, director: director}
}

func (c ConfigCmd) Run(opts ConfigOpts) error {

	if opts == (ConfigOpts{}) {
		return bosherr.Error("Either <ID> or parameters --type and --name must be provided")
	}

	if opts.Args.ID != "" && (opts.Type != "" || opts.Name != "") {
		return bosherr.Error("Can only specify one of ID or parameters '--type' and '--name'")
	}

	if (opts.Args.ID == "" && opts.Type != "" && opts.Name == "") || (opts.Args.ID == "" && opts.Name != "" && opts.Type == "") {
		return bosherr.Error("Need to specify both parameters '--type' and '--name'")
	}

	var config boshdir.Config
	var err error

	if opts.Args.ID != "" {
		config, err = c.director.LatestConfigById(opts.Args.ID)
	} else {
		config, err = c.director.LatestConfig(opts.Type, opts.Name)
	}

	if err != nil {
		return err
	}

	table := boshtbl.Table{
		Content: "config",

		Header: []boshtbl.Header{
			boshtbl.NewHeader("ID"),
			boshtbl.NewHeader("Type"),
			boshtbl.NewHeader("Name"),
			boshtbl.NewHeader("Created at"),
			boshtbl.NewHeader("Content"),
		},

		Notes: []string{},

		FillFirstColumn: true,

		Transpose: true,
	}

	table.Rows = append(table.Rows, []boshtbl.Value{
		boshtbl.NewValueString(config.ID),
		boshtbl.NewValueString(config.Type),
		boshtbl.NewValueString(config.Name),
		boshtbl.NewValueString(config.CreatedAt),
		boshtbl.NewValueString(config.Content),
	})

	c.ui.PrintTable(table)
	return nil
}
