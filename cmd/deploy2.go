package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-init/director"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type Deploy2Cmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewDeploy2Cmd(ui boshui.UI, deployment boshdir.Deployment) Deploy2Cmd {
	return Deploy2Cmd{ui: ui, deployment: deployment}
}

func (c Deploy2Cmd) Run(opts DeployOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables())
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating manifest")
	}

	man, err := boshdir.NewManifestFromBytes(bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking manifest")
	}

	if man.Name != c.deployment.Name() {
		errMsg := "Expected manifest to specify deployment name '%s' but was '%s'"
		return bosherr.Errorf(errMsg, c.deployment.Name(), man.Name)
	}

	diff, err := c.deployment.Diff(bytes, opts.NoRedact)
	if err != nil {
		return bosherr.WrapError(err, "Diffing manifest")
	}

	for _, line := range diff {
		lineMod, _ := line[1].(string)

		if lineMod == "added" {
			c.ui.BeginLinef("+ %s\n", line[0])
		} else if lineMod == "removed" {
			c.ui.BeginLinef("- %s\n", line[0])
		} else {
			c.ui.BeginLinef("  %s\n", line[0])
		}
	}

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	return c.deployment.Update(bytes, opts.Recreate, opts.SkipDrain)
}
