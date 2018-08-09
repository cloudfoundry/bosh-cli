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

	teams, err := deployment.Teams()
	if err != nil {
		return err
	}

	releases, err := deployment.Releases()
	if err != nil {
		return err
	}

	stemcells, err := deployment.Stemcells()
	if err != nil {
		return err
	}

	configs, err := c.director.ListDeploymentConfigs(deployment.Name())
	if err != nil {
		return err
	}

	return DeploymentTablePrinter{deployment.Name(), releases, stemcells, teams, configs, c.ui}.Print()
}
