package cmd

import (
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeleteReleaseCmd struct {
	ui       boshui.UI
	director boshdir.Director
}

func NewDeleteReleaseCmd(ui boshui.UI, director boshdir.Director) DeleteReleaseCmd {
	return DeleteReleaseCmd{ui: ui, director: director}
}

func (c DeleteReleaseCmd) Run(opts DeleteReleaseOpts) error {
	err := c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	releaseSlug, ok := opts.Args.Slug.ReleaseSlug()
	if ok {
		release, err := c.director.FindRelease(releaseSlug)
		if err != nil {
			return err
		}

		exists, err := release.Exists()
		if err != nil {
			return err
		}
		if !exists {
			c.ui.PrintLinef("Release '%s/%s' does not exist.", release.Name(), release.Version())
			return nil
		}

		return release.Delete(opts.Force)
	}

	releaseSeries, err := c.director.FindReleaseSeries(opts.Args.Slug.SeriesSlug())
	if err != nil {
		return err
	}

	exists, err := releaseSeries.Exists()
	if err != nil {
		return err
	}

	if !exists {
		c.ui.PrintLinef("Release '%s' does not exist.", releaseSeries.Name())
		return nil
	}

	return releaseSeries.Delete(opts.Force)
}
