package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type InspectLocalReleaseCmd struct {
	reader boshrel.Reader
	ui     biui.UI
}

func NewInspectLocalReleaseCmd(
	reader boshrel.Reader,
	ui biui.UI,
) InspectLocalReleaseCmd {
	return InspectLocalReleaseCmd{
		reader: reader,
		ui:     ui,
	}
}

func (cmd InspectLocalReleaseCmd) Run(opts InspectLocalReleaseOpts) error {
	release, err := cmd.reader.Read(opts.Args.PathToRelease)

	if err != nil {
		return err
	}
	defer release.CleanUp() //nolint:errcheck

	ReleaseTables{Release: release, ArchivePath: opts.Args.PathToRelease}.Print(cmd.ui)

	return nil
}
