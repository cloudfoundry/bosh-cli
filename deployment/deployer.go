package deployment

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstance "github.com/cloudfoundry/bosh-micro-cli/deployment/instance"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Deployer interface {
	Deploy(
		bmcloud.Cloud,
		bmdeplmanifest.Manifest,
		bmstemcell.CloudStemcell,
		bminstallmanifest.Registry,
		bminstallmanifest.SSHTunnel,
		bmvm.Manager,
		bmblobstore.Blobstore,
		bmui.Stage,
	) (Deployment, error)
}

type deployer struct {
	vmManagerFactory       bmvm.ManagerFactory
	instanceManagerFactory bminstance.ManagerFactory
	deploymentFactory      Factory
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	deploymentFactory Factory,
	logger boshlog.Logger,
) *deployer {
	return &deployer{
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		deploymentFactory:      deploymentFactory,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (d *deployer) Deploy(
	cloud bmcloud.Cloud,
	deploymentManifest bmdeplmanifest.Manifest,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	vmManager bmvm.Manager,
	blobstore bmblobstore.Blobstore,
	deployStage bmui.Stage,
) (Deployment, error) {
	instanceManager := d.instanceManagerFactory.NewManager(cloud, vmManager, blobstore)

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond
	if err := instanceManager.DeleteAll(pingTimeout, pingDelay, deployStage); err != nil {
		return nil, err
	}

	instances, disks, err := d.createAllInstances(deploymentManifest, instanceManager, cloudStemcell, registryConfig, sshTunnelConfig, deployStage)
	if err != nil {
		return nil, err
	}

	stemcells := []bmstemcell.CloudStemcell{cloudStemcell}
	return d.deploymentFactory.NewDeployment(instances, disks, stemcells), nil
}

func (d *deployer) createAllInstances(
	deploymentManifest bmdeplmanifest.Manifest,
	instanceManager bminstance.Manager,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	deployStage bmui.Stage,
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
			instance, instanceDisks, err := instanceManager.Create(jobSpec.Name, instanceID, deploymentManifest, cloudStemcell, registryConfig, sshTunnelConfig, deployStage)
			if err != nil {
				return instances, disks, bosherr.WrapErrorf(err, "Creating instance '%s/%d'", jobSpec.Name, instanceID)
			}
			instances = append(instances, instance)
			disks = append(disks, instanceDisks...)

			err = instance.UpdateJobs(deploymentManifest, deployStage)
			if err != nil {
				return instances, disks, err
			}
		}
	}

	return instances, disks, nil
}
