package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeploymentCmd struct {
	sessionFactory func(cmdconf.Config) Session

	director boshdir.Director
	config   cmdconf.Config
	ui       biui.UI
}

func NewDeploymentCmd(
	sessionFactory func(cmdconf.Config) Session,
	config cmdconf.Config,
	ui biui.UI,
	director boshdir.Director,
) DeploymentCmd {
	return DeploymentCmd{sessionFactory: sessionFactory, config: config, ui: ui, director: director}
}

func (c DeploymentCmd) Run() error {
	sess := c.sessionFactory(c.config)

	deployment, err := sess.Deployment()
	if err != nil {
		return err
	}

	// check if deployment exists
	_, err = deployment.Teams()
	if err != nil {
		return err
	}

	_, _, err = c.director.ListDeploymentConfigs(deployment.Name())
	if err != nil {
		return err
	}

	//      show configs instead of cloud configs
	return DeploymentTable{[]boshdir.Deployment{deployment}, c.ui}.Print()
}
