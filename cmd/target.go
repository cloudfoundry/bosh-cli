package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type TargetCmd struct {
	sessionFactory func(cmdconf.Config) Session

	config cmdconf.Config
	ui     boshui.UI
}

func NewTargetCmd(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui boshui.UI,
) TargetCmd {
	return TargetCmd{sessionFactory: sessionFactory, config: config, ui: ui}
}

func (c TargetCmd) Run(opts TargetOpts) error {
	args := opts.Args

	if len(args.URL) == 0 {
		return c.show()
	}

	updatedConfig := c.config.SetTarget(args.URL, args.Alias, opts.CACert.Path)

	err := c.set(updatedConfig)
	if err != nil {
		// If CA cert is specified, fail immediately
		if len(opts.CACert.Path) > 0 {
			return err
		}

		// Otherwise try existing CA cert if user is just switching between targets
		existingCACert := c.config.CACert(c.config.ResolveTarget(args.URL))

		updatedConfig = c.config.SetTarget(args.URL, args.Alias, existingCACert)

		altErr := c.set(updatedConfig)
		if altErr != nil {
			// Return original error without CA cert
			return err
		}
	}

	return nil
}

func (c TargetCmd) show() error {
	sess := c.sessionFactory(c.config)

	c.ui.PrintLinef("Current target is '%s'", sess.Target())

	director, err := sess.Director()
	if err != nil {
		return err
	}

	info, err := director.Info()
	if err != nil {
		return err
	}

	InfoTable{info, c.ui}.Print()

	return nil
}

func (c TargetCmd) set(updatedConfig cmdconf.Config) error {
	sess := c.sessionFactory(updatedConfig)

	director, err := sess.Director()
	if err != nil {
		return err
	}

	info, err := director.Info()
	if err != nil {
		return err
	}

	err = updatedConfig.Save()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Target set to '%s'", sess.Target())

	InfoTable{info, c.ui}.Print()

	return nil
}
