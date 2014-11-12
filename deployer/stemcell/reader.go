package stemcell

import (
	"encoding/json"
	"path/filepath"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

// Reader reads a stemcell tarball and returns a stemcell object containing
// parsed information (e.g. version, name)
//
type Reader interface {
	Read(stemcellTarballPath string, extractedPath string) (ExtractedStemcell, error)
}

type reader struct {
	compressor boshcmd.Compressor
	fs         boshsys.FileSystem
}

func NewReader(compressor boshcmd.Compressor, fs boshsys.FileSystem) Reader {
	return reader{compressor: compressor, fs: fs}
}

func (s reader) Read(stemcellTarballPath string, extractedPath string) (ExtractedStemcell, error) {
	err := s.compressor.DecompressFileToDir(stemcellTarballPath, extractedPath, boshcmd.CompressorOptions{})
	if err != nil {
		return nil, bosherr.WrapError(err, "Extracting stemcell from %s to %s", stemcellTarballPath, extractedPath)
	}

	var stemcellManifest Manifest
	stemcellManifestPath := filepath.Join(extractedPath, "stemcell.MF")

	stemcellManifestContents, err := s.fs.ReadFile(stemcellManifestPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading stemcell manifest %s", stemcellManifestPath)
	}

	err = candiedyaml.Unmarshal(stemcellManifestContents, &stemcellManifest)
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing stemcell manifest %s", stemcellManifestContents)
	}

	var stemcellApplySpec ApplySpec
	stemcellApplySpecPath := filepath.Join(extractedPath, "apply_spec.yml")

	stemcellApplySpecContents, err := s.fs.ReadFile(stemcellApplySpecPath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading stemcell apply spec %s", stemcellApplySpecPath)
	}

	err = json.Unmarshal(stemcellApplySpecContents, &stemcellApplySpec)
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing stemcell apply spec %s", stemcellApplySpecContents)
	}

	stemcellManifest.ImagePath = filepath.Join(extractedPath, "image")
	stemcell := NewExtractedStemcell(
		stemcellManifest,
		stemcellApplySpec,
		extractedPath,
		s.fs,
	)

	return stemcell, nil
}
