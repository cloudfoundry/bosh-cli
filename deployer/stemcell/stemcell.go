package stemcell

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type ExtractedStemcell interface {
	Manifest() Manifest
	ApplySpec() ApplySpec
	Delete() error
}

type extractedStemcell struct {
	manifest      Manifest
	applySpec     ApplySpec
	extractedPath string
	fs            boshsys.FileSystem
}

func NewExtractedStemcell(
	manifest Manifest,
	applySpec ApplySpec,
	extractedPath string,
	fs boshsys.FileSystem,
) ExtractedStemcell {
	return &extractedStemcell{
		manifest:      manifest,
		applySpec:     applySpec,
		extractedPath: extractedPath,
		fs:            fs,
	}
}

func (s *extractedStemcell) Manifest() Manifest { return s.manifest }

func (s *extractedStemcell) ApplySpec() ApplySpec { return s.applySpec }

func (s *extractedStemcell) Delete() error {
	return s.fs.RemoveAll(s.extractedPath)
}

type Manifest struct {
	ImagePath          string
	Name               string
	Version            string
	SHA1               string
	RawCloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

type ApplySpec struct {
	Job      Job
	Packages map[string]Blob
	Networks map[string]interface{}
}

type Job struct {
	Name      string
	Templates []Blob
}

type Blob struct {
	Name        string
	Version     string
	SHA1        string
	BlobstoreID string `json:"blobstore_id"`
}

func (m Manifest) CloudProperties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(m.RawCloudProperties)
}
