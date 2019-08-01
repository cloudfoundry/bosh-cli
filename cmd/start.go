package cmd

import (
	"errors"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type StartCmd struct {
	ui         boshui.UI
	deployment boshdir.Deployment
}

func NewStartCmd(ui boshui.UI, deployment boshdir.Deployment) StartCmd {
	return StartCmd{ui: ui, deployment: deployment}
}

func (c StartCmd) Run(opts StartOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	startOpts, err := NewStartOpts(opts)
	if err != nil {
		return err
	}

	err = validateSlug(opts.Args.Slug, startOpts)
	if err != nil {
		return err
	}
	return c.deployment.Start(opts.Args.Slug, startOpts)
}

func NewStartOpts(opts StartOpts) (boshdir.StartOpts, error) {
	if !opts.NoConverge { // converge is default, no-converge is opt-in
		startOpts := boshdir.StartOpts{
			Canaries:    opts.Canaries,
			MaxInFlight: opts.MaxInFlight,
			Converge:    true,
		}
		return startOpts, nil
	}

	if opts.Converge {
		return boshdir.StartOpts{}, errors.New("Can't set converge and no-converge")
	}

	if opts.Canaries != "" {
		return boshdir.StartOpts{}, errors.New("Can't set canaries and no-converge")
	}

	if opts.MaxInFlight != "" {
		return boshdir.StartOpts{}, errors.New("Can't set max-in-flight and no-converge")
	}

	return boshdir.StartOpts{Converge: false}, nil

}

func validateSlug(slug boshdir.AllOrInstanceGroupOrInstanceSlug, opts boshdir.StartOpts) error {
	if opts.Converge {
		return nil
	}
	if _, ok := slug.InstanceSlug(); !ok {
		return errors.New("You are trying to run start with --no-converge on an entire instance group. This operation is not allowed. Trying using the --converge flag or running it against a specific instance.")
	}
	return nil
}
