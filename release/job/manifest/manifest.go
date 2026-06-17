package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Name       string                        `yaml:"name"`
	Templates  map[string]string             `yaml:"templates"`
	Packages   []string                      `yaml:"packages"`
	Properties map[string]PropertyDefinition `yaml:"properties"`
	Consumes   []LinkDef                     `yaml:"consumes"`
	Provides   []LinkDef                     `yaml:"provides"`
}

type PropertyDefinition struct {
	Description string      `yaml:"description"`
	Default     interface{} `yaml:"default"`
}

// LinkDef represents a consumes or provides entry in a release job's spec file.
type LinkDef struct {
	Name       string   `yaml:"name"`
	Type       string   `yaml:"type"`
	Optional   bool     `yaml:"optional"`
	Properties []string `yaml:"properties"` // only meaningful on provides
}

func NewManifestFromPath(path string, fs boshsys.FileSystem) (Manifest, error) {
	var manifest Manifest

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return manifest, bosherr.WrapErrorf(err, "Reading job spec '%s'", path)
	}

	err = yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, bosherr.WrapErrorf(err, "Unmarshalling job spec '%s'", path)
	}

	return manifest, nil
}
