package stemcell

import (
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
)

type Packer interface {
	Pack(extractedStemcell ExtractedStemcell) (string, error)
}

type packer struct {
	compressor boshcmd.Compressor
}

func NewPacker(
	compressor boshcmd.Compressor,
) Packer {
	return &packer{
		compressor: compressor,
	}
}

func (p *packer) Pack(extractedStemcell ExtractedStemcell) (string, error) {
	err := extractedStemcell.Save()
	if err != nil {
		return "", err
	}

	tarballDestinationPath, err := p.compressor.CompressFilesInDir(extractedStemcell.GetExtractedPath())
	if err != nil {
		return "", err
	}

	err = extractedStemcell.Delete()
	if err != nil {
		return "", err
	}

	return tarballDestinationPath, nil
}
