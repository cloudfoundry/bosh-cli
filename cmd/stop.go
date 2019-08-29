package cmd

import (
	"errors"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type StopCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewStopCmd(ui boshui.UI, deployment boshdir.Deployment) StopCmd {
	return StopCmd{ui: ui, deployment: deployment}
}

func (c StopCmd) Run(opts StopOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	stopOpts, err := newStopOpts(opts)
	if err != nil {
		return err
	}

	return c.deployment.Stop(opts.Args.Slug, stopOpts)
}

func newStopOpts(opts StopOpts) (boshdir.StopOpts, error) {
	if !opts.NoConverge { // converge is default, no-converge is opt-in
		stopOpts := boshdir.StopOpts{
			Canaries:    opts.Canaries,
			MaxInFlight: opts.MaxInFlight,
			Hard:        opts.Hard,
			SkipDrain:   opts.SkipDrain,
			Converge:    true,
		}
		return stopOpts, nil
	}

	if opts.Converge {
		return boshdir.StopOpts{}, errors.New("Can't set converge and no-converge")
	}

	if opts.Canaries != "" {
		return boshdir.StopOpts{}, errors.New("Can't set canaries and no-converge")
	}

	if opts.MaxInFlight != "" {
		return boshdir.StopOpts{}, errors.New("Can't set max-in-flight and no-converge")
	}

	if _, ok := opts.Args.Slug.InstanceSlug(); !ok {
		return boshdir.StopOpts{}, errors.New("You are trying to run stop with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance.")
	}

	return boshdir.StopOpts{
		Converge:  false,
		Hard:      opts.Hard,
		SkipDrain: opts.SkipDrain,
	}, nil
}
