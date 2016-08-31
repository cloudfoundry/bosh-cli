package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type UploadReleaseCmd struct {
	releaseReader        boshrel.Reader
	releaseArchiveWriter boshrel.Writer
	releaseDir           boshreldir.ReleaseDir

	director              boshdir.Director
	releaseArchiveFactory func(string) boshdir.ReleaseArchive

	ui boshui.UI
}

func NewUploadReleaseCmd(
	releaseReader boshrel.Reader,
	releaseArchiveWriter boshrel.Writer,
	releaseDir boshreldir.ReleaseDir,
	director boshdir.Director,
	releaseArchiveFactory func(string) boshdir.ReleaseArchive,
	ui boshui.UI,
) UploadReleaseCmd {
	return UploadReleaseCmd{
		releaseReader:        releaseReader,
		releaseArchiveWriter: releaseArchiveWriter,
		releaseDir:           releaseDir,

		director:              director,
		releaseArchiveFactory: releaseArchiveFactory,

		ui: ui,
	}
}

func (c UploadReleaseCmd) Run(opts UploadReleaseOpts) error {
	if opts.Args.URL.IsRemote() {
		return c.uploadRemote(string(opts.Args.URL), opts)
	}

	if c.releaseReader == nil {
		return bosherr.Errorf("Cannot upload non-remote release '%s'", opts.Args.URL)
	}

	return c.uploadFile(opts.Args.URL.FilePath(), opts.Rebase, opts.Fix)
}

func (c UploadReleaseCmd) uploadRemote(url string, opts UploadReleaseOpts) error {
	version := semver.Version(opts.Version)

	necessary, err := c.needToUpload(opts.Name, version.AsString(), opts.Fix)
	if err != nil || !necessary {
		return err
	}

	return c.director.UploadReleaseURL(url, opts.SHA1, opts.Rebase, opts.Fix)
}

func (c UploadReleaseCmd) uploadFile(path string, rebase, fix bool) error {
	var release boshrel.Release
	var err error

	if len(path) > 0 {
		release, err = c.releaseReader.Read(path)
		if err != nil {
			return err
		}
	} else {
		release, err = c.releaseDir.LastRelease()
		if err != nil {
			return err
		}
	}

	var pkgFpsToSkip []string

	if !fix {
		pkgFpsToSkip, err = c.director.MatchPackages(release.Manifest(), release.IsCompiled())
		if err != nil {
			return err
		}
	}

	path, err = c.releaseArchiveWriter.Write(release, pkgFpsToSkip)
	if err != nil {
		return err
	}

	file, err := c.releaseArchiveFactory(path).File()
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening release")
	}

	return c.director.UploadReleaseFile(file, rebase, fix)
}

func (c UploadReleaseCmd) needToUpload(name, version string, fix bool) (bool, error) {
	if fix {
		return true, nil
	}

	found, err := c.director.HasRelease(name, version)
	if err != nil {
		return true, err
	}

	if found {
		c.ui.PrintLinef("Release '%s/%s' already exists.", name, version)
		return false, nil
	}

	return true, nil
}
