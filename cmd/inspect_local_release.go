package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	biui "github.com/cloudfoundry/bosh-cli/ui"
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

	ReleaseTables{Release: release, ArchivePath: opts.Args.PathToRelease}.Print(cmd.ui)

	return nil
}
