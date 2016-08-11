package cmd

import (
	"github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type BuildManifestCmd struct {
	ui boshui.UI
}

func NewBuildManifestCmd(ui boshui.UI) BuildManifestCmd {
	return BuildManifestCmd{
		ui: ui,
	}
}

func (c BuildManifestCmd) Run(opts BuildManifestOpts) error {
	variables := opts.VarFlags.AsVariables()

	template := template.NewTemplate(opts.Args.Manifest.Bytes)

	evaluatedManifest, err := template.Evaluate(variables)
	if err != nil {
		return err
	}

	c.ui.PrintBlock(string(evaluatedManifest.Content()))
	return nil
}
