package stemcell

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Reader interface {
	Read(stemcellPath string, extractedPath string) (Stemcell, error)
}

type reader struct {
	compressor boshcmd.Compressor
	fs         boshsys.FileSystem
}

func NewReader(compressor boshcmd.Compressor, fs boshsys.FileSystem) Reader {
	return reader{compressor: compressor, fs: fs}
}

func (s reader) Read(stemcellPath string, extractedPath string) (Stemcell, error) {
	err := s.compressor.DecompressFileToDir(stemcellPath, extractedPath)
	if err != nil {
		return Stemcell{}, bosherr.WrapError(err, "Extracting stemcell from %s to %s", stemcellPath, extractedPath)
	}

	var stemcell Stemcell
	stemcellManifestPath := filepath.Join(extractedPath, "stemcell.MF")
	stemcellContents, err := s.fs.ReadFile(stemcellManifestPath)
	if err != nil {
		return Stemcell{}, bosherr.WrapError(err, "Reading stemcell manifest %s", stemcellManifestPath)
	}

	err = candiedyaml.Unmarshal(stemcellContents, &stemcell)
	if err != nil {
		return Stemcell{}, bosherr.WrapError(err, "Parsing stemcell manifest %s", stemcellContents)
	}

	stemcell.ImagePath = filepath.Join(extractedPath, "image")

	return stemcell, nil
}
