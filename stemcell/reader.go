package stemcell

import (
	"path/filepath"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

type manifest struct {
	Name            string
	Version         string
	SHA1            string
	CloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

// Reader reads a stemcell tarball and returns a stemcell object containing
// parsed information (e.g. version, name)
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
		return nil, bosherr.WrapErrorf(err, "Extracting stemcell from '%s' to '%s'", stemcellTarballPath, extractedPath)
	}

	var rawManifest manifest
	manifestPath := filepath.Join(extractedPath, "stemcell.MF")

	manifestContents, err := s.fs.ReadFile(manifestPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading stemcell manifest '%s'", manifestPath)
	}

	err = candiedyaml.Unmarshal(manifestContents, &rawManifest)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing stemcell manifest: %s", manifestContents)
	}

	manifest := Manifest{
		Name:    rawManifest.Name,
		Version: rawManifest.Version,
		SHA1:    rawManifest.SHA1,
	}

	cloudProperties, err := biproperty.BuildMap(rawManifest.CloudProperties)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing stemcell cloud_properties: %#v", rawManifest.CloudProperties)
	}
	manifest.CloudProperties = cloudProperties

	manifest.ImagePath = filepath.Join(extractedPath, "image")

	stemcell := NewExtractedStemcell(
		manifest,
		extractedPath,
		s.fs,
	)

	return stemcell, nil
}
