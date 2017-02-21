package stemcell

import (
	"fmt"

	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	yaml "gopkg.in/yaml.v2"
)

type ExtractedStemcell interface {
	Manifest() Manifest
	Delete() error
	OsAndVersion() string
	SetName(string)
	SetVersion(string)
	SetCloudProperties(string) error
	GetExtractedPath() string
	Save() error
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

func (s *extractedStemcell) OsAndVersion() string {
	return fmt.Sprintf("%s/%s", s.manifest.OS, s.manifest.Version)
}

func (s *extractedStemcell) SetName(newName string) {
	s.manifest.Name = newName
}

func (s *extractedStemcell) SetVersion(newVersion string) {
	s.manifest.Version = newVersion
}

func (s *extractedStemcell) SetCloudProperties(newCloudProperties string) error {
	newProps := new(biproperty.Map)

	err := yaml.Unmarshal([]byte(newCloudProperties), newProps)

	for key, value := range *newProps {
		s.manifest.CloudProperties[key] = value
	}

	return err
}

func (s *extractedStemcell) GetExtractedPath() string {
	return s.extractedPath
}

func (s *extractedStemcell) Save() error {
	// TODO(cdutra): implement me
	return nil
}

type Manifest struct {
	ImagePath       string
	Name            string
	Version         string
	OS              string
	SHA1            string
	CloudProperties biproperty.Map
}
