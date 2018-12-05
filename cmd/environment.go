package cmd

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type EnvironmentCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewEnvironmentCmd(ui boshui.UI, director boshdir.Director) EnvironmentCmd {
	return EnvironmentCmd{ui: ui, director: director}
}

func (c EnvironmentCmd) Run(opts EnvironmentOpts) error {
	info, err := c.director.Info()
	if err != nil {
		return err
	}

	InfoTable{info, c.ui}.Print()

	if opts.Details {
		certificatesInfo, err := c.director.CertificateExpiry()

		if err != nil {
			return err
		}

		CertificateInfoTable{certificatesInfo, c.ui}.Print()
	}

	return nil
}
