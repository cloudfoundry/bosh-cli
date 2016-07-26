package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-init/cmd/config"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type EnvironmentCmd struct {
	sessionFactory func(cmdconf.Config) Session

	config cmdconf.Config
	ui     boshui.UI
}

func NewEnvironmentCmd(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui boshui.UI,
) EnvironmentCmd {
	return EnvironmentCmd{sessionFactory: sessionFactory, config: config, ui: ui}
}

func (c EnvironmentCmd) Run(opts EnvironmentOpts) error {
	args := opts.Args

	if len(args.URL) == 0 {
		return c.show()
	}

	updatedConfig := c.config.SetEnvironment(args.URL, args.Alias, opts.CACert.Path)

	err := c.set(updatedConfig)
	if err != nil {
		// If CA cert is specified, fail immediately
		if len(opts.CACert.Path) > 0 {
			return err
		}

		// Otherwise try existing CA cert if user is just switching between targets
		existingCACert := c.config.CACert(c.config.ResolveEnvironment(args.URL))

		updatedConfig = c.config.SetEnvironment(args.URL, args.Alias, existingCACert)

		altErr := c.set(updatedConfig)
		if altErr != nil {
			// Return original error without CA cert
			return err
		}
	}

	return nil
}

func (c EnvironmentCmd) show() error {
	sess := c.sessionFactory(c.config)

	c.ui.PrintLinef("Current target is '%s'", sess.Environment())

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

func (c EnvironmentCmd) set(updatedConfig cmdconf.Config) error {
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

	c.ui.PrintLinef("Target set to '%s'", sess.Environment())

	InfoTable{info, c.ui}.Print()

	return nil
}
