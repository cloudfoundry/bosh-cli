package cmd

import (
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type BuildManifestCmd struct {
	ui boshui.UI
}

func NewBuildManifestCmd(ui boshui.UI) BuildManifestCmd {
	return BuildManifestCmd{ui: ui}
}

func (c BuildManifestCmd) Run(opts BuildManifestOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOps())
	if err != nil {
		return err
	}

	c.ui.PrintBlock(string(bytes))

	return nil
}
