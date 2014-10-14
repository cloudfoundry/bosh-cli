package deployment

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type boshDeploymentParser struct {
	fs boshsys.FileSystem
}

func NewBoshDeploymentParser(fs boshsys.FileSystem) ManifestParser {
	return boshDeploymentParser{fs: fs}
}

type boshDeploymentManifest struct {
	Name          string
	Networks      []Network
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	Jobs          []Job
}

func (p boshDeploymentParser) Parse(path string) (Deployment, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Deployment{}, bosherr.WrapError(err, "Reading file %s", path)
	}

	depManifest := boshDeploymentManifest{}
	err = candiedyaml.Unmarshal(contents, &depManifest)
	if err != nil {
		return Deployment{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}

	deployment := Deployment{
		Name:          depManifest.Name,
		Networks:      depManifest.Networks,
		ResourcePools: depManifest.ResourcePools,
		Jobs:          depManifest.Jobs,
	}

	return deployment, nil
}
