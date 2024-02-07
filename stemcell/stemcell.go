package stemcell

import (
	"fmt"
	"path/filepath"

	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"gopkg.in/yaml.v2"
)

type ExtractedStemcell interface {
	Manifest() Manifest
	Cleanup() error
	OsAndVersion() string
	SetName(string)
	SetVersion(string)
	SetFormat([]string)
	SetCloudProperties(biproperty.Map)
	GetExtractedPath() string
	Pack(string) error
	EmptyImage() error
	fmt.Stringer
}

type extractedStemcell struct {
	manifest      Manifest
	extractedPath string
	compressor    boshfu.Compressor
	fs            boshsys.FileSystem
}

type Manifest struct {
	Name            string         `yaml:"name"`
	Version         string         `yaml:"version"`
	OS              string         `yaml:"operating_system"`
	SHA1            string         `yaml:"sha1"`
	BoshProtocol    string         `yaml:"bosh_protocol"`
	StemcellFormats []string       `yaml:"stemcell_formats,omitempty"`
	ApiVersion      int            `yaml:"api_version,omitempty"`
	CloudProperties biproperty.Map `yaml:"cloud_properties"`
}

func NewExtractedStemcell(
	manifest Manifest,
	extractedPath string,
	compressor boshfu.Compressor,
	fs boshsys.FileSystem,
) ExtractedStemcell {
	return &extractedStemcell{
		manifest:      manifest,
		extractedPath: extractedPath,
		compressor:    compressor,
		fs:            fs,
	}
}

func (s *extractedStemcell) Manifest() Manifest { return s.manifest }

func (s *extractedStemcell) Cleanup() error {
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

func (s *extractedStemcell) SetFormat(newFormats []string) {
	s.manifest.StemcellFormats = newFormats
}

func (s *extractedStemcell) SetCloudProperties(newCloudProperties biproperty.Map) {
	for key, value := range newCloudProperties {
		s.manifest.CloudProperties[key] = value
	}
}

func (s *extractedStemcell) Pack(destinationPath string) error {
	defer s.Cleanup() //nolint:errcheck

	err := s.save()
	if err != nil {
		return err
	}

	paths, err := s.fs.Glob(filepath.Join(s.extractedPath, "*"))
	if err != nil {
		return err
	}

	filenames := []string{}
	for _, p := range paths {
		filenames = append(filenames, filepath.Base(p))
	}

	intermediateStemcellPath, err := s.compressor.CompressSpecificFilesInDir(s.extractedPath, filenames)
	if err != nil {
		return err
	}

	return boshfu.NewFileMover(s.fs).Move(intermediateStemcellPath, destinationPath)
}

func (s *extractedStemcell) EmptyImage() error {
	imagePath := filepath.Join(s.extractedPath, "image")
	err := s.fs.WriteFile(imagePath, []byte{})
	if err != nil {
		return err
	}
	return nil
}

func (s *extractedStemcell) GetExtractedPath() string {
	return s.extractedPath
}

func (s *extractedStemcell) save() error {
	stemcellMfPath := filepath.Join(s.extractedPath, "stemcell.MF")
	contents, _ := yaml.Marshal(s.manifest)
	err := s.fs.WriteFile(stemcellMfPath, contents)
	if err != nil {
		return err
	}
	return nil
}
