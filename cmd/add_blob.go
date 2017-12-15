package cmd

import (
	"os"

	"github.com/cloudfoundry/bosh-cli/crypto"
	"github.com/cloudfoundry/bosh-cli/installation/tarball"
	boshreldir "github.com/cloudfoundry/bosh-cli/releasedir"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type AddBlobCmd struct {
	blobsDir         boshreldir.BlobsDir
	tarballProvider  tarball.Provider
	digestCalculator crypto.DigestCalculator
	fs               boshsys.FileSystem
	ui               boshui.UI
}

func NewAddBlobCmd(blobsDir boshreldir.BlobsDir, tarballProvider tarball.Provider, digestCalculator crypto.DigestCalculator, fs boshsys.FileSystem, ui boshui.UI) AddBlobCmd {
	return AddBlobCmd{
		blobsDir:         blobsDir,
		tarballProvider:  tarballProvider,
		digestCalculator: digestCalculator,
		fs:               fs,
		ui:               ui,
	}
}

func (c AddBlobCmd) Run(stage boshui.Stage, opts AddBlobOpts) error {
	digest := opts.SHA1

	path, err := c.tarballProvider.Get(blobSource{URL: opts.Args.Path, SHA1: digest}, stage)
	if err != nil {
		panic(err)
	}

	if digest == "" {
		digest, err = c.digestCalculator.Calculate(path)
		if err != nil {
			return bosherr.WrapErrorf(err, "Calculating temp blob sha1")
		}
	}

	file, err := c.fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapErrorf(err, "Opening blob")
	}

	defer file.Close()

	blob, err := c.blobsDir.TrackBlob(path, digest, file)
	if err != nil {
		return bosherr.WrapErrorf(err, "Tracking blob")
	}

	c.ui.PrintLinef("Added blob '%s'", blob.Path)

	return nil
}

type blobSource struct {
	URL  string
	SHA1 string
}

func (bs blobSource) GetURL() string      { return bs.URL }
func (bs blobSource) GetSHA1() string     { return bs.SHA1 }
func (bs blobSource) Description() string { return "no description" }
