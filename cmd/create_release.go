package cmd

import (
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshrel "github.com/cloudfoundry/bosh-init/release"
	boshreldir "github.com/cloudfoundry/bosh-init/releasedir"
	boshui "github.com/cloudfoundry/bosh-init/ui"
)

type CreateReleaseCmd struct {
	releaseManifestReader boshrel.Reader
	releaseDir            boshreldir.ReleaseDir
	ui                    boshui.UI
}

func NewCreateReleaseCmd(
	releaseManifestReader boshrel.Reader,
	releaseDir boshreldir.ReleaseDir,
	ui boshui.UI,
) CreateReleaseCmd {
	return CreateReleaseCmd{
		releaseManifestReader: releaseManifestReader,
		releaseDir:            releaseDir,
		ui:                    ui,
	}
}

func (c CreateReleaseCmd) Run(opts CreateReleaseOpts) error {
	var release boshrel.Release
	var err error

	manifestGiven := len(opts.Args.Manifest.Path) > 0

	if manifestGiven {
		release, err = c.releaseManifestReader.Read(opts.Args.Manifest.Path)
		if err != nil {
			return err
		}
	} else {
		release, err = c.buildRelease(opts)
		if err != nil {
			return err
		}

		if opts.Final {
			err = c.finalizeRelease(opts, release)
			if err != nil {
				return err
			}
		}
	}

	var archivePath string

	if manifestGiven || opts.WithTarball {
		archivePath, err = c.releaseDir.BuildReleaseArchive(release)
		if err != nil {
			return err
		}
	}

	ReleaseTables{Release: release, ArchivePath: archivePath}.Print(c.ui)

	return nil
}

func (c CreateReleaseCmd) buildRelease(opts CreateReleaseOpts) (boshrel.Release, error) {
	var err error

	name := opts.Name

	if len(name) == 0 {
		name, err = c.releaseDir.DefaultName()
		if err != nil {
			return nil, err
		}
	}

	version := semver.Version(opts.Version)

	if version.Empty() {
		version, err = c.releaseDir.NextDevVersion(name, opts.TimestampVersion)
		if err != nil {
			return nil, err
		}
	}

	return c.releaseDir.BuildRelease(name, version, opts.Force)
}

func (c CreateReleaseCmd) finalizeRelease(opts CreateReleaseOpts, release boshrel.Release) error {
	version := semver.Version(opts.Version)

	if version.Empty() {
		version, err := c.releaseDir.NextFinalVersion(release.Name())
		if err != nil {
			return err
		}

		release.SetVersion(version.AsString())
	}

	return c.releaseDir.FinalizeRelease(release, opts.Force)
}
