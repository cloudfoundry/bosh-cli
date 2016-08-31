package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeploymentCmd struct {
	sessionFactory func(cmdconf.Config) Session

	config cmdconf.Config
	ui     biui.UI
}

func NewDeploymentCmd(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui biui.UI,
) DeploymentCmd {
	return DeploymentCmd{sessionFactory: sessionFactory, config: config, ui: ui}
}

func (c DeploymentCmd) Run(opts DeploymentOpts) error {
	if len(opts.Args.NameOrPath) == 0 {
		return c.show()
	}
	return c.set(opts)
}

func (c DeploymentCmd) show() error {
	sess := c.sessionFactory(c.config)

	deployment, err := sess.Deployment()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Current deployment is '%s'", deployment.Name())

	_ = DeploymentsTable{[]boshdir.Deployment{deployment}, c.ui}.Print()

	return nil
}

func (c DeploymentCmd) set(opts DeploymentOpts) error {
	sess := c.sessionFactory(c.config)

	updatedConfig := c.config.SetDeployment(sess.Environment(), opts.Args.NameOrPath)

	sess = c.sessionFactory(updatedConfig)

	deployment, err := sess.Deployment()
	if err != nil {
		return err
	}

	err = updatedConfig.Save()
	if err != nil {
		return err
	}

	c.ui.PrintLinef("Deployment set to '%s'", deployment.Name())

	_ = DeploymentsTable{[]boshdir.Deployment{deployment}, c.ui}.Print()

	return nil
}
