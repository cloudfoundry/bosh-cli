package cmd

import (
	cmdconf "github.com/cloudfoundry/bosh-cli/cmd/config"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeploymentCmd struct {
	session Session

	config cmdconf.Config
	ui     biui.UI
}

func NewDeploymentCmd(
	session Session,
	config cmdconf.Config,
	ui biui.UI,
) DeploymentCmd {
	return DeploymentCmd{session: session, config: config, ui: ui}
}

func (c DeploymentCmd) Run() error {
	deployment, err := c.session.Deployment()
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

	director, err := c.session.Director()
	if err != nil {
		return err
	}

	configs, err := director.ListDeploymentConfigs(deployment.Name())
	if err != nil {
		return err
	}

	return DeploymentTablePrinter{deployment.Name(), releases, stemcells, teams, configs, c.ui}.Print()
}
