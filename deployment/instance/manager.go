package instance

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/blobstore"
	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Manager interface {
	FindCurrent() ([]Instance, error)
	Create(
		jobName string,
		id int,
		deploymentManifest bmdeplmanifest.Manifest,
		cloudStemcell bmstemcell.CloudStemcell,
		registryConfig bminstallmanifest.Registry,
		sshTunnelConfig bminstallmanifest.SSHTunnel,
		eventLoggerStage bmui.Stage,
	) (Instance, []bmdisk.Disk, error)
	DeleteAll(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		eventLoggerStage bmui.Stage,
	) error
}

type manager struct {
	cloud            bmcloud.Cloud
	vmManager        bmvm.Manager
	blobstore        bmblobstore.Blobstore
	sshTunnelFactory bmsshtunnel.Factory
	instanceFactory  Factory
	logger           boshlog.Logger
	logTag           string
}

func NewManager(
	cloud bmcloud.Cloud,
	vmManager bmvm.Manager,
	blobstore bmblobstore.Blobstore,
	sshTunnelFactory bmsshtunnel.Factory,
	instanceFactory Factory,
	logger boshlog.Logger,
) Manager {
	return &manager{
		cloud:            cloud,
		vmManager:        vmManager,
		blobstore:        blobstore,
		sshTunnelFactory: sshTunnelFactory,
		instanceFactory:  instanceFactory,
		logger:           logger,
		logTag:           "vmDeployer",
	}
}

func (m *manager) FindCurrent() ([]Instance, error) {
	instances := []Instance{}

	// Only one current instance will exist (for now)
	vm, found, err := m.vmManager.FindCurrent()
	if err != nil {
		return instances, bosherr.WrapError(err, "Finding currently deployed instances")
	}

	if found {
		// TODO: store the name of the job for each instance in the repo, so that we can print it when deleting
		jobName := "unknown"
		instanceID := 0

		instance := m.instanceFactory.NewInstance(
			jobName,
			instanceID,
			vm,
			m.vmManager,
			m.sshTunnelFactory,
			m.blobstore,
			m.logger,
		)
		instances = append(instances, instance)
	}

	return instances, nil
}

func (m *manager) Create(
	jobName string,
	id int,
	deploymentManifest bmdeplmanifest.Manifest,
	cloudStemcell bmstemcell.CloudStemcell,
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	eventLoggerStage bmui.Stage,
) (Instance, []bmdisk.Disk, error) {
	var vm bmvm.VM
	stepName := fmt.Sprintf("Creating VM for instance '%s/%d' from stemcell '%s'", jobName, id, cloudStemcell.CID())
	err := eventLoggerStage.Perform(stepName, func() error {
		var err error
		vm, err = m.vmManager.Create(cloudStemcell, deploymentManifest)
		if err != nil {
			return bosherr.WrapError(err, "Creating VM")
		}

		if err = cloudStemcell.PromoteAsCurrent(); err != nil {
			return bosherr.WrapErrorf(err, "Promoting stemcell as current '%s'", cloudStemcell.CID())
		}

		return nil
	})
	if err != nil {
		return nil, []bmdisk.Disk{}, err
	}

	instance := m.instanceFactory.NewInstance(jobName, id, vm, m.vmManager, m.sshTunnelFactory, m.blobstore, m.logger)

	if err := instance.WaitUntilReady(registryConfig, sshTunnelConfig, eventLoggerStage); err != nil {
		return instance, []bmdisk.Disk{}, bosherr.WrapError(err, "Waiting until instance is ready")
	}

	disks, err := instance.UpdateDisks(deploymentManifest, eventLoggerStage)
	if err != nil {
		return instance, disks, bosherr.WrapError(err, "Updating instance disks")
	}

	return instance, disks, err
}

func (m *manager) DeleteAll(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	eventLoggerStage bmui.Stage,
) error {
	instances, err := m.FindCurrent()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		if err = instance.Delete(pingTimeout, pingDelay, eventLoggerStage); err != nil {
			return bosherr.WrapErrorf(err, "Deleting existing instance '%s/%d'", instance.JobName(), instance.ID())
		}
	}
	return nil
}
