package deployment

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type CPIDeploymentParser interface {
	Parse(path string) (CPIDeployment, error)
}

type microDeploymentParser struct {
	fs boshsys.FileSystem
}

func NewCpiDeploymentParser(fs boshsys.FileSystem) CPIDeploymentParser {
	return microDeploymentParser{fs: fs}
}

type cpiDeploymentManifest struct {
	Name          string
	CloudProvider cloudProviderProperties `yaml:"cloud_provider"`
}

type cloudProviderProperties struct {
	Properties      map[interface{}]interface{}
	Registry        Registry
	AgentEnvService string    `yaml:"agent_env_service"`
	SSHTunnel       SSHTunnel `yaml:"ssh_tunnel"`
	Mbus            string
}

func (m microDeploymentParser) Parse(path string) (CPIDeployment, error) {
	contents, err := m.fs.ReadFile(path)
	if err != nil {
		return CPIDeployment{}, bosherr.WrapError(err, "Reading file %s", path)
	}

	depManifest := cpiDeploymentManifest{}
	err = candiedyaml.Unmarshal(contents, &depManifest)
	if err != nil {
		return CPIDeployment{}, bosherr.WrapError(err, "Parsing job manifest")
	}

	deployment := CPIDeployment{
		Name:            depManifest.Name,
		Jobs:            m.defaultCPIJobs(),
		Registry:        depManifest.CloudProvider.Registry,
		AgentEnvService: depManifest.CloudProvider.AgentEnvService,
		SSHTunnel:       depManifest.CloudProvider.SSHTunnel,
		Mbus:            depManifest.CloudProvider.Mbus,
		RawProperties:   depManifest.CloudProvider.Properties,
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
