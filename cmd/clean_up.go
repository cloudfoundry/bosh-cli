package cmd

import (
	"fmt"

	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type CleanUpCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewCleanUpCmd(ui boshui.UI, director boshdir.Director) CleanUpCmd {
	return CleanUpCmd{ui: ui, director: director}
}

func (c CleanUpCmd) Run(opts CleanUpOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	resp, err := c.director.CleanUp(opts.All, opts.DryRun)
	if err != nil {
		return err
	}
	return printCleanUpTable(resp)
}

func printCleanUpTable(resp boshdir.CleanUp) error {
	fmt.Println("PRINTING TABLE!!!")
	return nil

}
