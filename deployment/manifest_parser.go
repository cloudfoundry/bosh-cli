package deployment

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type ManifestParser interface {
	Parse(manifestPath string) (LocalDeployment, error)
}

type manifestParser struct {
	fs boshsys.FileSystem
}

func NewManifestParser(fs boshsys.FileSystem) ManifestParser {
	return manifestParser{fs: fs}
}

type LocalDeployment struct {
	Name       string
	Properties map[string]interface{}
}

type deploymentManifest struct {
	Name          string
	CloudProvider cloudProviderProperties `yaml:"cloud_provider"`
}

type cloudProviderProperties struct {
	Properties map[string]interface{}
}

func (m manifestParser) Parse(path string) (LocalDeployment, error) {
	contents, err := m.fs.ReadFile(path)
	if err != nil {
		return LocalDeployment{}, bosherr.WrapError(err, "Reading file %s", path)
	}

	depManifest := deploymentManifest{}
	err = candiedyaml.Unmarshal(contents, &depManifest)
	if err != nil {
		return LocalDeployment{}, bosherr.WrapError(err, "Parsing job manifest")
	}

	localDeployment := LocalDeployment{
		Name:       depManifest.Name,
		Properties: depManifest.CloudProvider.Properties,
	}

	return localDeployment, nil
}
