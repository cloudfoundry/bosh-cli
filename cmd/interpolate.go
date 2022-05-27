package cmd

import (
	"github.com/cppforlife/go-patch/patch"

	. "github.com/cloudfoundry/bosh-cli/v6/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/v6/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/v6/ui"
)

type InterpolateCmd struct {
	ui boshui.UI
}

func NewInterpolateCmd(ui boshui.UI) InterpolateCmd {
	return InterpolateCmd{ui: ui}
}

func (c InterpolateCmd) Run(opts InterpolateOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	vars := opts.VarFlags.AsVariables()
	op := opts.OpsFlags.AsOp()
	evalOpts := boshtpl.EvaluateOpts{
		ExpectAllKeys:     opts.VarErrors,
		ExpectAllVarsUsed: opts.VarErrorsUnused,
	}

	if opts.Path.IsSet() {
		evalOpts.PostVarSubstitutionOp = patch.FindOp{Path: opts.Path}

		// Printing YAML indented multiline strings (eg SSH key) is not useful
		evalOpts.UnescapedMultiline = true
	}

	bytes, err := tpl.Evaluate(vars, op, evalOpts)
	if err != nil {
		return err
	}

	c.ui.PrintBlock(bytes)

	return nil
}
