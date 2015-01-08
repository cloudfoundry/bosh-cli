package manifest

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Parser interface {
	Parse(path string) (Manifest, error)
}

type parser struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

type manifest struct {
	Name          string
	Releases      []ReleaseRef
	Update        UpdateSpec
	Networks      []Network
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	DiskPools     []DiskPool     `yaml:"disk_pools"`
	Jobs          []Job
}

type UpdateSpec struct {
	UpdateWatchTime *string `yaml:"update_watch_time"`
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

func (p *parser) Parse(path string) (Manifest, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	comboManifest := manifest{}
	err = candiedyaml.Unmarshal(contents, &comboManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}
	p.logger.Debug(p.logTag, "Parsed BOSH deployment manifest: %#v", comboManifest)

	deploymentManifest, err := p.parseDeploymentManifest(comboManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}

	return deploymentManifest, nil
}

func (p *parser) parseDeploymentManifest(depManifest manifest) (Manifest, error) {
	deployment := boshDeploymentDefaults
	deployment.Name = depManifest.Name
	deployment.Releases = depManifest.Releases
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
