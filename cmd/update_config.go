package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type UpdateConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewUpdateConfigCmd(ui boshui.UI, director boshdir.Director) UpdateConfigCmd {
	return UpdateConfigCmd{ui: ui, director: director}
}

func (c UpdateConfigCmd) Run(opts UpdateConfigOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Config.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating config")
	}

	configDiff, err := c.director.DiffConfig(opts.Args.Type, opts.Name, bytes)
	if err != nil {
		return err
	}

	config, err := c.director.LatestConfig(opts.Args.Type, opts.Name)
	if err != nil && err.Error() != "No config" {
		return err
	}

	diff := NewDiff(configDiff.Diff)
	if (config != boshdir.Config{} && len(diff.lines) == 0) {
		c.ui.PrintLinef("no changes in config, nothing to update\n")
		return nil
	}
	diff.Print(c.ui)

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.UpdateConfig(opts.Args.Type, opts.Name, bytes)
}
