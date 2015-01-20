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
		cloud bmcloud.Cloud,
		deploymentManifest bmdeplmanifest.Manifest,
		extractedStemcell bmstemcell.ExtractedStemcell,
		registryConfig bminstallmanifest.Registry,
		sshTunnelConfig bminstallmanifest.SSHTunnel,
		vmManager bmvm.Manager,
		blobstoreURL string,
	) (Deployment, error)
}

type deployer struct {
	stemcellManagerFactory bmstemcell.ManagerFactory
	vmManagerFactory       bmvm.ManagerFactory
	instanceManagerFactory bminstance.ManagerFactory
	deploymentFactory      Factory
	eventLogger            bmeventlog.EventLogger
	logger                 boshlog.Logger
	logTag                 string
}

func NewDeployer(
	stemcellManagerFactory bmstemcell.ManagerFactory,
	vmManagerFactory bmvm.ManagerFactory,
	instanceManagerFactory bminstance.ManagerFactory,
	deploymentFactory Factory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	return &deployer{
		stemcellManagerFactory: stemcellManagerFactory,
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		deploymentFactory:      deploymentFactory,
		eventLogger:            eventLogger,
		logger:                 logger,
		logTag:                 "deployer",
	}
}

func (d *deployer) Deploy(
	cloud bmcloud.Cloud,
	deploymentManifest bmdeplmanifest.Manifest,
	extractedStemcell bmstemcell.ExtractedStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	vmManager bmvm.Manager,
	blobstoreURL string,
) (Deployment, error) {

	//TODO: handle stage construction outside of this class
	uploadStemcellStage := d.eventLogger.NewStage("uploading stemcell")
	uploadStemcellStage.Start()

	stemcellManager := d.stemcellManagerFactory.NewManager(cloud)
	cloudStemcell, err := stemcellManager.Upload(extractedStemcell, uploadStemcellStage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Uploading stemcell")
	}
	stemcells := []bmstemcell.CloudStemcell{cloudStemcell}

	uploadStemcellStage.Finish()

	deployStage := d.eventLogger.NewStage("deploying")
	deployStage.Start()

	instanceManager := d.instanceManagerFactory.NewManager(cloud, vmManager, blobstoreURL)

	pingTimeout := 10 * time.Second
	pingDelay := 500 * time.Millisecond
	if err = instanceManager.DeleteAll(pingTimeout, pingDelay, deployStage); err != nil {
		return nil, err
	}

	instances, disks, err := d.createAllInstances(deploymentManifest, instanceManager, extractedStemcell, cloudStemcell, registryConfig, sshTunnelConfig, deployStage)
	if err != nil {
		return nil, err
	}

	// TODO: cleanup unused disks?

	if err = stemcellManager.DeleteUnused(deployStage); err != nil {
		return nil, err
	}

	deployStage.Finish()

	return d.deploymentFactory.NewDeployment(instances, disks, stemcells), nil
}

func (d *deployer) createAllInstances(
	deploymentManifest bmdeplmanifest.Manifest,
	instanceManager bminstance.Manager,
	extractedStemcell bmstemcell.ExtractedStemcell,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	deployStage bmeventlog.Stage,
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

			//TODO: compile packages (on the vm), upload compiled packages (to the blobstore)
			// instance.CompilePackages

			err = instance.UpdateJobs(deploymentManifest, extractedStemcell.ApplySpec(), deployStage)
			if err != nil {
				return instances, disks, err
			}
		}
	}

	return instances, disks, nil
}
