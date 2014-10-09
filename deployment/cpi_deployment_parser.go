package deployment

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type microDeploymentParser struct {
	fs boshsys.FileSystem
}

func NewCpiDeploymentParser(fs boshsys.FileSystem) ManifestParser {
	return microDeploymentParser{fs: fs}
}

type cpiDeploymentManifest struct {
	Name          string
	CloudProvider cloudProviderProperties `yaml:"cloud_provider"`
}

type cloudProviderProperties struct {
	Properties map[interface{}]interface{}
	Registry   Registry
}

func (m microDeploymentParser) Parse(path string) (Deployment, error) {
	contents, err := m.fs.ReadFile(path)
	if err != nil {
		return Deployment{}, bosherr.WrapError(err, "Reading file %s", path)
	}

	depManifest := cpiDeploymentManifest{}
	err = candiedyaml.Unmarshal(contents, &depManifest)
	if err != nil {
		return Deployment{}, bosherr.WrapError(err, "Parsing job manifest")
	}

	properties, err := bmkeystr.NewKeyStringifier().ConvertMap(depManifest.CloudProvider.Properties)
	if err != nil {
		return Deployment{}, bosherr.WrapError(err, "Converting manifest cloud properties")
	}

	deployment := Deployment{
		Name:       depManifest.Name,
		Properties: properties,
		Jobs:       m.defaultCPIJobs(),
		Registry:   depManifest.CloudProvider.Registry,
	}

	return deployment, nil
}

func (m microDeploymentParser) defaultCPIJobs() []Job {
	return []Job{
		Job{
			Name:      "cpi",
			Instances: 1,
			Templates: []ReleaseJobRef{
				ReleaseJobRef{
					Name:    "cpi",
					Release: "unknown-cpi-release-name",
				},
			},
		},
	}
}
