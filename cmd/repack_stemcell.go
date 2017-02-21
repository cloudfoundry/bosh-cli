package cmd

import (
	"errors"
	"github.com/cloudfoundry/bosh-cli/stemcell"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type RepackStemcellCmd struct {
	ui                boshui.UI
	fs                boshsys.FileSystem
	stemcellExtractor stemcell.Extractor
	stemcellPacker    stemcell.Packer
}

func NewRepackStemcellCmd(
	ui boshui.UI,
	fs boshsys.FileSystem,
	stemcellExtractor stemcell.Extractor,
	stemcellPacker stemcell.Packer,
) RepackStemcellCmd {
	return RepackStemcellCmd{ui: ui, fs: fs, stemcellExtractor: stemcellExtractor, stemcellPacker: stemcellPacker}
}

func (c RepackStemcellCmd) Run(opts RepackStemcellOpts) error {
	if c.fs.FileExists(opts.Args.PathToResult) {
		return errors.New("destination file exists")
	}

	extractedStemcell, err := c.stemcellExtractor.Extract(opts.Args.PathToStemcell)
	if err != nil {
		return err
	}
	intermediateStemcellPath, err := c.stemcellPacker.Pack(extractedStemcell)
	if err != nil {
		return err
	}

	return c.fs.Rename(intermediateStemcellPath, opts.Args.PathToResult)
}
