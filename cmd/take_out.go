package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cloudfoundry/bosh-cli/takeout"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
	"os"
)

type TakeOutCmd struct {
	ui boshui.UI
	to takeout.Utensils
}

func NewTakeOutCmd(ui boshui.UI, d takeout.Utensils) TakeOutCmd {
	return TakeOutCmd{ui: ui, to: d}
}

func (c TakeOutCmd) Run(opts TakeOutOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating manifest")
	}
	if _, err := os.Stat(opts.Args.Name); os.IsExist(err) {
		return bosherr.WrapErrorf(err, "Takeout op name exists")
	}

	manifest, err := boshdir.NewManifestFromBytes(bytes)
	c.ui.PrintLinef("Processing releases for offline use")
	var releaseChanges []takeout.OpEntry
	for _, r := range manifest.Releases {
		if r.URL == "" {
			c.ui.PrintLinef("Release does not have a URL for take_out; Name: %s / %s", r.Name, r.Version)
			return bosherr.WrapErrorf(nil, "Provide an opsfile that has a URL or removes this release") // TODO
		} else {
			o, err := c.to.TakeOutRelease(r, c.ui, opts.MirrorPrefix)
			if err != nil {
				return err
			}
			releaseChanges = append(releaseChanges, o)
		}
	}

	y, _ := yaml.Marshal(releaseChanges)
	c.ui.PrintLinef("Writing take_out operation to file: " + opts.Args.Name)
	takeoutOp, err := os.Create(opts.Args.Name)
	if err != nil {
		return err
	}
	defer takeoutOp.Close()
	takeoutOp.WriteString("---\n")
	takeoutOp.WriteString(string(y))
	return nil
}
