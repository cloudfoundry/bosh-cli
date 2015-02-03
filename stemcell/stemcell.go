package stemcell

import (
	"fmt"

	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type ExtractedStemcell interface {
	Manifest() Manifest
	ApplySpec() ApplySpec
	Delete() error
	fmt.Stringer
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

func (s *extractedStemcell) String() string {
	return fmt.Sprintf("ExtractedStemcell{name=%s version=%s}", s.manifest.Name, s.manifest.Version)
}

type Manifest struct {
	ImagePath       string
	Name            string
	Version         string
	SHA1            string
	CloudProperties bmproperty.Map
}

type ApplySpec struct {
	Job      Job
	Packages map[string]Blob
	Networks map[string]bmproperty.Map
}

type Job struct {
	Name      string
	Templates []Blob
}

type Blob struct {
	Name        string
	Version     string
	SHA1        string
	BlobstoreID string
}
