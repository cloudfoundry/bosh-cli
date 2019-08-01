package cmd

import (
	"errors"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type RestartCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewRestartCmd(ui boshui.UI, deployment boshdir.Deployment) RestartCmd {
	return RestartCmd{ui: ui, deployment: deployment}
}

func (c RestartCmd) Run(opts RestartOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	restartOpts, err := newRestartOpts(opts)
	if err != nil {
		return err
	}
	return c.deployment.Restart(opts.Args.Slug, restartOpts)
}

func newRestartOpts(opts RestartOpts) (boshdir.RestartOpts, error) {
	if !opts.NoConverge { // converge is default, no-converge is opt-in
		restartOpts := boshdir.RestartOpts{
			Canaries:    opts.Canaries,
			MaxInFlight: opts.MaxInFlight,
			SkipDrain:   opts.SkipDrain,
			Converge:    true,
		}
		return restartOpts, nil
	}

	if opts.Converge {
		return boshdir.RestartOpts{}, errors.New("Can't set converge and no-converge")
	}

	if opts.Canaries != "" {
		return boshdir.RestartOpts{}, errors.New("Can't set canaries and no-converge")
	}

	if opts.MaxInFlight != "" {
		return boshdir.RestartOpts{}, errors.New("Can't set max-in-flight and no-converge")
	}

	if _, ok := opts.Args.Slug.InstanceSlug(); !ok {
		return boshdir.RestartOpts{}, errors.New("You are trying to run restart with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance.")
	}

	return boshdir.RestartOpts{
		Converge:  false,
		SkipDrain: opts.SkipDrain,
	}, nil
}
