package cmd

import (
	"errors"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type RecreateCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewRecreateCmd(ui boshui.UI, deployment boshdir.Deployment) RecreateCmd {
	return RecreateCmd{ui: ui, deployment: deployment}
}

func (c RecreateCmd) Run(opts RecreateOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	recreateOpts, err := newRecreateOpts(opts)
	if err != nil {
		return err
	}
	return c.deployment.Recreate(opts.Args.Slug, recreateOpts)
}

func newRecreateOpts(opts RecreateOpts) (boshdir.RecreateOpts, error) {
	if !opts.NoConverge { // converge is default, no-converge is opt-in
		recreateOpts := boshdir.RecreateOpts{
			SkipDrain:   opts.SkipDrain,
			Fix:         opts.Fix,
			DryRun:      opts.DryRun,
			Canaries:    opts.Canaries,
			MaxInFlight: opts.MaxInFlight,
			Converge:    true,
		}
		return recreateOpts, nil
	}

	if opts.Converge {
		return boshdir.RecreateOpts{}, errors.New("Can't set converge and no-converge")
	}

	if opts.Canaries != "" {
		return boshdir.RecreateOpts{}, errors.New("Can't set canaries and no-converge")
	}

	if opts.MaxInFlight != "" {
		return boshdir.RecreateOpts{}, errors.New("Can't set max-in-flight and no-converge")
	}

	if opts.DryRun {
		return boshdir.RecreateOpts{}, errors.New("Can't set dry-run and no-converge")
	}

	if _, ok := opts.Args.Slug.InstanceSlug(); !ok {
		return boshdir.RecreateOpts{}, errors.New("You are trying to run recreate with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance.")
	}

	return boshdir.RecreateOpts{
		Converge:  false,
		SkipDrain: opts.SkipDrain,
		Fix:       opts.Fix,
	}, nil
}
