package stemcell

import (
	"fmt"

	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
)

type ExtractedStemcell interface {
	Manifest() Manifest
	Delete() error
	fmt.Stringer
}

type extractedStemcell struct {
	manifest      Manifest
	extractedPath string
	fs            boshsys.FileSystem
}

func NewExtractedStemcell(
	manifest Manifest,
	extractedPath string,
	fs boshsys.FileSystem,
) ExtractedStemcell {
	return &extractedStemcell{
		manifest:      manifest,
		extractedPath: extractedPath,
		fs:            fs,
	}
}

func (s *extractedStemcell) Manifest() Manifest { return s.manifest }

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
