package stemcell

import (
	"encoding/json"
	"path/filepath"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type manifest struct {
	Name            string
	Version         string
	SHA1            string
	CloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

type applySpec struct {
	Job      job
	Packages map[string]blob
	Networks map[string]map[interface{}]interface{}
}

type job struct {
	Name      string
	Templates []blob
}

type blob struct {
	Name        string
	Version     string
	SHA1        string
	BlobstoreID string `json:"blobstore_id"`
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

	var rawApplySpec applySpec
	applySpecPath := filepath.Join(extractedPath, "apply_spec.yml")

	applySpecContents, err := s.fs.ReadFile(applySpecPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading stemcell apply spec '%s'", applySpecPath)
	}

	err = json.Unmarshal(applySpecContents, &rawApplySpec)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing stemcell apply spec: %s", applySpecContents)
	}

	manifest := Manifest{
		Name:    rawManifest.Name,
		Version: rawManifest.Version,
		SHA1:    rawManifest.SHA1,
	}

	cloudProperties, err := bmproperty.BuildMap(rawManifest.CloudProperties)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Parsing stemcell cloud_properties: %#v", rawManifest.CloudProperties)
	}
	manifest.CloudProperties = cloudProperties

	manifest.ImagePath = filepath.Join(extractedPath, "image")

	applySpec := ApplySpec{
		Job: Job{
			Name: rawApplySpec.Job.Name,
		},
	}

	jobTemplates := make([]Blob, len(rawApplySpec.Job.Templates), len(rawApplySpec.Job.Templates))
	for i, rawJobTemplate := range rawApplySpec.Job.Templates {
		jobTemplates[i] = Blob{
			Name:        rawJobTemplate.Name,
			Version:     rawJobTemplate.Version,
			SHA1:        rawJobTemplate.SHA1,
			BlobstoreID: rawJobTemplate.BlobstoreID,
		}
	}
	applySpec.Job.Templates = jobTemplates

	if rawApplySpec.Packages != nil {
		packages := make(map[string]Blob, len(rawApplySpec.Packages))
		for packageName, rawPackage := range rawApplySpec.Packages {
			packages[packageName] = Blob{
				Name:        rawPackage.Name,
				Version:     rawPackage.Version,
				SHA1:        rawPackage.SHA1,
				BlobstoreID: rawPackage.BlobstoreID,
			}
		}
		applySpec.Packages = packages
	}

	if rawApplySpec.Networks != nil {
		networks := make(map[string]bmproperty.Map, len(rawApplySpec.Networks))
		for networkName, rawNetworkInterface := range rawApplySpec.Networks {
			networkInterface, err := bmproperty.BuildMap(rawNetworkInterface)
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Parsing stemcell network '%s' interface: %#v", networkName, rawNetworkInterface)
			}
			networks[networkName] = networkInterface
		}
		applySpec.Networks = networks
	}

	stemcell := NewExtractedStemcell(
		manifest,
		applySpec,
		extractedPath,
		s.fs,
	)

	return stemcell, nil
}
