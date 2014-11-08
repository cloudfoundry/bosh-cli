package deployment

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type boshDeploymentParser struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

func NewBoshDeploymentParser(fs boshsys.FileSystem, logger boshlog.Logger) ManifestParser {
	return boshDeploymentParser{
		fs:     fs,
		logger: logger,
		logTag: "boshDeploymentParser",
	}
}

type boshDeploymentManifest struct {
	Name          string
	Update        UpdateSpec
	Networks      []Network
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	DiskPools     []DiskPool     `yaml:"disk_pools"`
	Jobs          []Job
}

type UpdateSpec struct {
	UpdateWatchTime *string `yaml:"update_watch_time"`
}

var boshDeploymentDefaults = Deployment{
	Update: Update{
		UpdateWatchTime: WatchTime{
			Start: 0,
			End:   300000,
		},
	},
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

	p.logger.Debug(p.logTag, "Parsed BOSH deployment manifest: %#v", depManifest)

	deployment := boshDeploymentDefaults
	deployment.Name = depManifest.Name
	deployment.Networks = depManifest.Networks
	deployment.ResourcePools = depManifest.ResourcePools
	deployment.DiskPools = depManifest.DiskPools
	deployment.Jobs = depManifest.Jobs

	if depManifest.Update.UpdateWatchTime != nil {
		updateWatchTime, err := NewWatchTime(*depManifest.Update.UpdateWatchTime)
		if err != nil {
			return Deployment{}, bosherr.WrapError(err, "Parsing update watch time")
		}

		deployment.Update = Update{
			UpdateWatchTime: updateWatchTime,
		}
	}

	return deployment, nil
}
