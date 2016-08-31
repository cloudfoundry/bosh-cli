package cmd

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type RunErrandCmd struct {
	deployment boshdir.Deployment
	downloader Downloader
	ui         biui.UI
}

func NewRunErrandCmd(
	deployment boshdir.Deployment,
	downloader Downloader,
	ui biui.UI,
) RunErrandCmd {
	return RunErrandCmd{deployment: deployment, downloader: downloader, ui: ui}
}

func (c RunErrandCmd) Run(opts RunErrandOpts) error {
	result, err := c.deployment.RunErrand(opts.Args.Name, opts.KeepAlive)
	if err != nil {
		return err
	}

	errandErr := c.summarize(opts.Args.Name, result)

	if opts.DownloadLogs && len(result.LogsBlobstoreID) > 0 {
		err := c.downloader.Download(
			result.LogsBlobstoreID,
			result.LogsSHA1,
			opts.Args.Name,
			opts.LogsDirectory.Path,
		)
		if err != nil {
			return bosherr.WrapError(err, "Downloading errand logs")
		}
	}

	return errandErr
}

func (c RunErrandCmd) summarize(errandName string, result boshdir.ErrandResult) error {
	c.printOutput("[stdout]", result.Stdout)
	c.printOutput("[stderr]", result.Stderr)

	prefix := fmt.Sprintf("Errand '%s'", errandName)
	suffix := fmt.Sprintf("(exit code %d)", result.ExitCode)

	switch {
	case result.ExitCode == 0:
		c.ui.PrintLinef("%s completed successfully %s", prefix, suffix)
		return nil
	case result.ExitCode > 128:
		return bosherr.Errorf("%s was canceled %s", prefix, suffix)
	default:
		return bosherr.Errorf("%s completed with error %s", prefix, suffix)
	}
}

func (c RunErrandCmd) printOutput(title, output string) {
	if len(output) > 0 {
		c.ui.PrintLinef("%s", title)
		c.ui.PrintLinef("%s", output)
	}
}
