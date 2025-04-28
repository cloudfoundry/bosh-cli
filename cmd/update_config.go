package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
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

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{}) //nolint:staticcheck
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating config")
	}
	configDiff, err := c.director.DiffConfig(opts.Type, opts.Name, bytes)
	if err != nil {
		return err
	}

	diff := NewDiff(configDiff.Diff)
	diff.Print(c.ui)

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	var expectedId string
	if opts.ExpectedLatestId == "" {
		expectedId = configDiff.FromId
	} else {
		expectedId = opts.ExpectedLatestId
	}
	config, err := c.director.UpdateConfig(opts.Type, opts.Name, expectedId, bytes)
	if err != nil {
		return err
	}

	ConfigTable{config, c.ui}.Print()

	return nil
}
