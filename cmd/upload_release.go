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
	releaseDirFactory    func(DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir)
	releaseArchiveWriter boshrel.Writer

	director              boshdir.Director
	releaseArchiveFactory func(string) boshdir.ReleaseArchive

	ui boshui.UI
}

func NewUploadReleaseCmd(
	releaseDirFactory func(DirOrCWDArg) (boshrel.Reader, boshreldir.ReleaseDir),
	releaseArchiveWriter boshrel.Writer,
	director boshdir.Director,
	releaseArchiveFactory func(string) boshdir.ReleaseArchive,
	ui boshui.UI,
) UploadReleaseCmd {
	return UploadReleaseCmd{
		releaseDirFactory:    releaseDirFactory,
		releaseArchiveWriter: releaseArchiveWriter,

		director:              director,
		releaseArchiveFactory: releaseArchiveFactory,

		ui: ui,
	}
}

func (c UploadReleaseCmd) Run(opts UploadReleaseOpts) error {
	if opts.Release != nil {
		return c.uploadRelease(opts.Release, opts)
	}

	if opts.Args.URL.IsRemote() {
		return c.uploadRemote(string(opts.Args.URL), opts)
	}

	return c.uploadFile(opts)
}

func (c UploadReleaseCmd) uploadRemote(url string, opts UploadReleaseOpts) error {
	version := semver.Version(opts.Version)

	necessary, err := c.needToUpload(opts.Name, version.AsString(), opts.Fix)
	if err != nil || !necessary {
		return err
	}

	return c.director.UploadReleaseURL(url, opts.SHA1, opts.Rebase, opts.Fix)
}

func (c UploadReleaseCmd) uploadFile(opts UploadReleaseOpts) error {
	path := opts.Args.URL.FilePath()

	if c.releaseDirFactory == nil {
		return bosherr.Errorf("Cannot upload non-remote release '%s'", path)
	}

	releaseReader, releaseDir := c.releaseDirFactory(opts.Directory)

	var release boshrel.Release
	var err error

	if len(path) > 0 {
		release, err = releaseReader.Read(path)
		if err != nil {
			return err
		}
	} else {
		release, err = releaseDir.LastRelease()
		if err != nil {
			return err
		}
	}

	return c.uploadRelease(release, opts)
}

func (c UploadReleaseCmd) uploadRelease(release boshrel.Release, opts UploadReleaseOpts) error {
	var pkgFpsToSkip []string
	var err error

	if !opts.Fix {
		pkgFpsToSkip, err = c.director.MatchPackages(release.Manifest(), release.IsCompiled())
		if err != nil {
			return err
		}
	}

	path, err := c.releaseArchiveWriter.Write(release, pkgFpsToSkip)
	if err != nil {
		return err
	}

	file, err := c.releaseArchiveFactory(path).File()
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening release")
	}

	return c.director.UploadReleaseFile(file, opts.Rebase, opts.Fix)
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
