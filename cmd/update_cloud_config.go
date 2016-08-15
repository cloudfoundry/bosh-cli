package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type UpdateCloudConfigCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewUpdateCloudConfigCmd(ui boshui.UI, director boshdir.Director) UpdateCloudConfigCmd {
	return UpdateCloudConfigCmd{ui: ui, director: director}
}

func (c UpdateCloudConfigCmd) Run(opts UpdateCloudConfigOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.CloudConfig.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables())
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating cloud config")
	}

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.director.UpdateCloudConfig(bytes)
}
