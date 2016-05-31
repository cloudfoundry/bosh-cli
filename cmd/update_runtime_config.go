package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type UpdateRuntimeConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewUpdateRuntimeConfigCmd(ui boshui.UI, director boshdir.Director) UpdateRuntimeConfigCmd {
	return UpdateRuntimeConfigCmd{ui: ui, director: director}
}

func (c UpdateRuntimeConfigCmd) Run(opts UpdateRuntimeConfigOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.RuntimeConfig.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables())
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating runtime config")
	}

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.UpdateRuntimeConfig(bytes)
}
