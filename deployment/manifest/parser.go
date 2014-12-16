package manifest

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Parser interface {
	Parse(path string) (Manifest, CPIDeploymentManifest, error)
}

type parser struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

type manifest struct {
	Name          string
	Update        UpdateSpec
	Networks      []Network
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	DiskPools     []DiskPool     `yaml:"disk_pools"`
	Jobs          []Job
	CloudProvider cloudProviderProperties `yaml:"cloud_provider"`
}

type UpdateSpec struct {
	UpdateWatchTime *string `yaml:"update_watch_time"`
}

type cloudProviderProperties struct {
	Properties      map[interface{}]interface{}
	Registry        Registry
	AgentEnvService string    `yaml:"agent_env_service"`
	SSHTunnel       SSHTunnel `yaml:"ssh_tunnel"`
	Mbus            string
}

var boshDeploymentDefaults = Manifest{
	Update: Update{
		UpdateWatchTime: WatchTime{
			Start: 0,
			End:   300000,
		},
	},
}

func NewParser(fs boshsys.FileSystem, logger boshlog.Logger) Parser {
	return &parser{
		fs:     fs,
		logger: logger,
		logTag: "deploymentParser",
	}
}

func (p *parser) Parse(path string) (Manifest, CPIDeploymentManifest, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Manifest{}, CPIDeploymentManifest{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	depManifest := manifest{}
	err = candiedyaml.Unmarshal(contents, &depManifest)
	if err != nil {
		return Manifest{}, CPIDeploymentManifest{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}
	p.logger.Debug(p.logTag, "Parsed BOSH deployment manifest: %#v", depManifest)

	deployment, err := p.parseBoshDeployment(depManifest)
	if err != nil {
		return Manifest{}, CPIDeploymentManifest{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}

	cpiDeploymentManifest := p.parseCPIDeploymentManifest(depManifest)

	return deployment, cpiDeploymentManifest, nil
}

func (p *parser) parseBoshDeployment(depManifest manifest) (Manifest, error) {
	deployment := boshDeploymentDefaults
	deployment.Name = depManifest.Name
	deployment.Networks = depManifest.Networks
	deployment.ResourcePools = depManifest.ResourcePools
	deployment.DiskPools = depManifest.DiskPools
	deployment.Jobs = depManifest.Jobs

	if depManifest.Update.UpdateWatchTime != nil {
		updateWatchTime, err := NewWatchTime(*depManifest.Update.UpdateWatchTime)
		if err != nil {
			return Manifest{}, bosherr.WrapError(err, "Parsing update watch time")
		}

		deployment.Update = Update{
			UpdateWatchTime: updateWatchTime,
		}
	}

	return deployment, nil
}

func (p *parser) parseCPIDeploymentManifest(depManifest manifest) CPIDeploymentManifest {
	return CPIDeploymentManifest{
		Name:            depManifest.Name,
		Registry:        depManifest.CloudProvider.Registry,
		AgentEnvService: depManifest.CloudProvider.AgentEnvService,
		SSHTunnel:       depManifest.CloudProvider.SSHTunnel,
		Mbus:            depManifest.CloudProvider.Mbus,
		RawProperties:   depManifest.CloudProvider.Properties,
	}
}
