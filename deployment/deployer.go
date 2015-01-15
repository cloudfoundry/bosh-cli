package deployment

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

type Deployer interface {
	Deploy(
		bmcloud.Cloud,
		bmdeplmanifest.Manifest,
		bmstemcell.ExtractedStemcell,
		bminstallmanifest.Registry,
		bminstallmanifest.SSHTunnel,
		bmvm.Manager,
	) (Deployment, error)
}

type deployer struct {
	stemcellManagerFactory bmstemcell.ManagerFactory
	vmManagerFactory       bmvm.ManagerFactory
	instanceManagerFactory bminstance.ManagerFactory
	eventLoggerStage       bmeventlog.Stage
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	stemcellManagerFactory bmstemcell.ManagerFactory,
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	//TODO: handle stage construction outside of this class
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		stemcellManagerFactory: stemcellManagerFactory,
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		eventLoggerStage:       eventLoggerStage,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deploymentManifest bmdeplmanifest.Manifest,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	vmManager bmvm.Manager,
) (Deployment, error) {
	stemcellManager := m.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell)
	if err != nil {
		return nil, bosherr.WrapError(err, "Uploading stemcell")
	}
	stemcells := []bmstemcell.CloudStemcell{cloudStemcell}

	m.eventLoggerStage.Start()

	instanceManager := m.instanceManagerFactory.NewManager(cloud, vmManager)

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond
	if err = instanceManager.DeleteAll(pingTimeout, pingDelay, m.eventLoggerStage); err != nil {
		return nil, err
	}

	instances, disks, err := m.createAllInstances(deploymentManifest, instanceManager, extractedStemcell, cloudStemcell, registryConfig, sshTunnelConfig)
	if err != nil {
		return nil, err
	}

	// TODO: cleanup unused disks?

	if err = stemcellManager.DeleteUnused(m.eventLoggerStage); err != nil {
		return nil, err
	}

	m.eventLoggerStage.Finish()

	return NewDeployment(instances, disks, stemcells), nil
}

func (m *deployer) createAllInstances(
	deploymentManifest bmdeplmanifest.Manifest,
	instanceManager bminstance.Manager,
	extractedStemcell bmstemcell.ExtractedStemcell,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
) ([]bminstance.Instance, []bmdisk.Disk, error) {
	instances := []bminstance.Instance{}
	disks := []bmdisk.Disk{}

	if len(deploymentManifest.Jobs) != 1 {
		return instances, disks, bosherr.Errorf("There must only be one job, found %d", len(deploymentManifest.Jobs))
	}

	for _, jobSpec := range deploymentManifest.Jobs {
		if jobSpec.Instances != 1 {
			return instances, disks, bosherr.Errorf("Job '%s' must have only one instance, found %d", jobSpec.Name, jobSpec.Instances)
		}
		for instanceID := 0; instanceID < jobSpec.Instances; instanceID++ {
			instance, instanceDisks, err := instanceManager.Create(jobSpec.Name, instanceID, deploymentManifest, cloudStemcell, registryConfig, sshTunnelConfig, m.eventLoggerStage)
			if err != nil {
				return instances, disks, bosherr.WrapErrorf(err, "Creating instance '%s/%d'", jobSpec.Name, instanceID)
			}
			instances = append(instances, instance)
			disks = append(disks, instanceDisks...)

			err = instance.StartJobs(extractedStemcell.ApplySpec(), deploymentManifest, m.eventLoggerStage)
			if err != nil {
				return instances, disks, err
			}
		}
	}

	return instances, disks, nil
}
