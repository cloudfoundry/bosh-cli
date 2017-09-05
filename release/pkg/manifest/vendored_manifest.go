package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type VendoredManifest struct {
	Name         string   `yaml:"name"`
	Fingerprint  string   `yaml:"fingerprint"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

func NewVendoredManifestFromPath(path string, fs boshsys.FileSystem) (VendoredManifest, error) {
	var manifest VendoredManifest

	bytes, err := fs.ReadFile(path)
	if err != nil {
		return manifest, bosherr.WrapErrorf(err, "Reading package vendored spec '%s'", path)
	}

	err = yaml.Unmarshal(bytes, &manifest)
	if err != nil {
		return manifest, bosherr.WrapError(err, "Unmarshalling package vendored spec")
	}

	return manifest, nil
}

func (m VendoredManifest) AsBytes() ([]byte, error) {
	if len(m.Name) == 0 {
		return nil, bosherr.Errorf("Expected non-empty package name")
	}

	if len(m.Fingerprint) == 0 {
		return nil, bosherr.Errorf("Expected non-empty package fingerprint")
	}

	return yaml.Marshal(m)
}
