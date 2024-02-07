package cmd

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	"gopkg.in/yaml.v2"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	"github.com/cloudfoundry/bosh-cli/v7/stemcell"
)

type RepackStemcellCmd struct {
	stemcellExtractor stemcell.Extractor
}

func NewRepackStemcellCmd(
	stemcellExtractor stemcell.Extractor,
) RepackStemcellCmd {
	return RepackStemcellCmd{stemcellExtractor: stemcellExtractor}
}

func (c RepackStemcellCmd) Run(opts RepackStemcellOpts) error {
	extractedStemcell, err := c.stemcellExtractor.Extract(opts.Args.PathToStemcell)
	if err != nil {
		return err
	}

	if opts.Name != "" {
		extractedStemcell.SetName(opts.Name)
	}

	if opts.Version != "" {
		extractedStemcell.SetVersion(opts.Version)
	}

	if opts.EmptyImage {
		err = extractedStemcell.EmptyImage()
		if err != nil {
			return err
		}
	}

	if opts.CloudProperties != "" {
		cloudProperties := new(biproperty.Map)
		err = yaml.Unmarshal([]byte(opts.CloudProperties), cloudProperties)
		if err != nil {
			return err
		}

		extractedStemcell.SetCloudProperties(*cloudProperties)
	}

	if len(opts.Format) != 0 {
		extractedStemcell.SetFormat(opts.Format)
	}

	return extractedStemcell.Pack(opts.Args.PathToResult.ExpandedPath)
}
