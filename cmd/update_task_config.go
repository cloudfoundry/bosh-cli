package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type UpdateTaskConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewUpdateTaskConfigCmd(ui boshui.UI, director boshdir.Director) UpdateTaskConfigCmd {
	return UpdateTaskConfigCmd{ui: ui, director: director}
}

func (c UpdateTaskConfigCmd) Run(opts UpdateTaskConfigOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.TaskConfig.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating Task config")
	}

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.UpdateTaskConfig(bytes)
}
