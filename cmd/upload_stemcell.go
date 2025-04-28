package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type UploadStemcellCmd struct {
	director               boshdir.Director
	stemcellArchiveFactory func(string) boshdir.StemcellArchive

	ui biui.UI
}

func NewUploadStemcellCmd(
	director boshdir.Director,
	stemcellArchiveFactory func(string) boshdir.StemcellArchive,
	ui biui.UI,
) UploadStemcellCmd {
	return UploadStemcellCmd{
		director:               director,
		stemcellArchiveFactory: stemcellArchiveFactory,
		ui:                     ui,
	}
}

func (c UploadStemcellCmd) Run(opts UploadStemcellOpts) error {
	if opts.Args.URL.IsRemote() {
		return c.uploadRemote(string(opts.Args.URL), opts)
	}

	return c.uploadFile(opts.Args.URL.FilePath(), opts.Fix)
}

func (c UploadStemcellCmd) uploadRemote(url string, opts UploadStemcellOpts) error {
	version := semver.Version(opts.Version)

	necessary, err := c.needToUpload(opts.Name, version.AsString(), opts.Fix)
	if err != nil || !necessary {
		return err
	}

	return c.director.UploadStemcellURL(url, opts.SHA1, opts.Fix)
}

func (c UploadStemcellCmd) uploadFile(path string, fix bool) error {
	archive := c.stemcellArchiveFactory(path)

	stemcellMetadata, err := archive.Info()
	if err != nil {
		return bosherr.WrapErrorf(err, "Retrieving stemcell info")
	}

	necessary, err := c.needToUpload(stemcellMetadata.Name, stemcellMetadata.Version, fix)
	if err != nil || !necessary {
		return err
	}

	file, err := archive.File()
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening stemcell")
	}

	return c.director.UploadStemcellFile(file, fix)
}

func (c UploadStemcellCmd) needToUpload(name, version string, fix bool) (bool, error) {
	if fix {
		return true, nil
	}

	needed, err := c.director.StemcellNeedsUpload(
		boshdir.StemcellInfo{Name: name, Version: version},
	)
	if err != nil {
		return false, err
	}

	if !needed {
		c.ui.PrintLinef("Stemcell '%s/%s' already exists.", name, version)
		return false, nil
	}

	return true, nil
}
