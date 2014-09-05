package deployment

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type microDeploymentParser struct {
	fs boshsys.FileSystem
}

func NewMicroDeploymentParser(fs boshsys.FileSystem) ManifestParser {
	return microDeploymentParser{fs: fs}
}

type microDeploymentManifest struct {
	Name          string
	CloudProvider cloudProviderProperties `yaml:"cloud_provider"`
}

type cloudProviderProperties struct {
	Properties map[string]interface{}
}

func (m microDeploymentParser) Parse(path string) (Deployment, error) {
	contents, err := m.fs.ReadFile(path)
	if err != nil {
		return LocalDeployment{}, bosherr.WrapError(err, "Reading file %s", path)
	}

	depManifest := microDeploymentManifest{}
	err = candiedyaml.Unmarshal(contents, &depManifest)
	if err != nil {
		return LocalDeployment{}, bosherr.WrapError(err, "Parsing job manifest")
	}

	localDeployment := NewLocalDeployment(
		depManifest.Name,
		depManifest.CloudProvider.Properties,
	)

	return localDeployment, nil
}
