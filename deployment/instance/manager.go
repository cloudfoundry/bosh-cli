package instance

import (
	"fmt"
	"net/http"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	biblobstore "github.com/cloudfoundry/bosh-cli/v7/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/v7/deployment/disk"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/v7/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/v7/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type Manager interface {
	FindCurrent() ([]Instance, error)
	Create(
		jobName string,
		id int,
		deploymentManifest bideplmanifest.Manifest,
		cloudStemcell bistemcell.CloudStemcell,
		diskCIDs []string,
		eventLoggerStage biui.Stage,
	) (Instance, []bidisk.Disk, error)
	DeleteAll(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		skipDrain bool,
		eventLoggerStage biui.Stage,
	) error
}

type manager struct {
	cloud                bicloud.Cloud
	vmManager            bivm.Manager
	blobstoreFactory     biblobstore.Factory
	blobstoreHTTPClient  *http.Client
	sshTunnelFactory     bisshtunnel.Factory
	instanceFactory      Factory
	logger               boshlog.Logger
	logTag               string
}

func NewManager(
	cloud bicloud.Cloud,
	vmManager bivm.Manager,
	blobstoreFactory biblobstore.Factory,
	blobstoreHTTPClient *http.Client,
	sshTunnelFactory bisshtunnel.Factory,
	instanceFactory Factory,
	logger boshlog.Logger,
) Manager {
	return &manager{
		cloud:               cloud,
		vmManager:           vmManager,
		blobstoreFactory:    blobstoreFactory,
		blobstoreHTTPClient: blobstoreHTTPClient,
		sshTunnelFactory:    sshTunnelFactory,
		instanceFactory:     instanceFactory,
		logger:              logger,
		logTag:              "vmDeployer",
	}
}

func (m *manager) FindCurrent() ([]Instance, error) {
	instances := []Instance{}

	existingVMs, err := m.vmManager.FindAll()
	if err != nil {
		return instances, bosherr.WrapError(err, "Finding currently deployed instances")
	}

	for _, existing := range existingVMs {
		// For delete-only operations the stateBuilder inside the instance never
		// uses the blobstore, so a nil blobstore is acceptable when no factory
		// is available.
		var blobstore biblobstore.Blobstore
		if m.blobstoreFactory != nil {
			var err error
			blobstore, err = m.blobstoreFactory.Create(existing.VM.MbusURL(), m.blobstoreHTTPClient)
			if err != nil {
				return instances, bosherr.WrapErrorf(err, "Creating blobstore for instance '%s/%d'", existing.JobName, existing.InstanceID)
			}
		}
		instance := m.instanceFactory.NewInstance(
			existing.JobName,
			existing.InstanceID,
			existing.VM,
			m.vmManager,
			m.sshTunnelFactory,
			blobstore,
			m.logger,
		)
		instances = append(instances, instance)
	}

	return instances, nil
}

func (m *manager) Create(
	jobName string,
	id int,
	deploymentManifest bideplmanifest.Manifest,
	cloudStemcell bistemcell.CloudStemcell,
	diskCIDs []string,
	eventLoggerStage biui.Stage,
) (Instance, []bidisk.Disk, error) {
	var vm bivm.VM
	stepName := fmt.Sprintf("Creating VM for instance '%s/%d' from stemcell '%s'", jobName, id, cloudStemcell.CID())
	err := eventLoggerStage.Perform(stepName, func() error {
		var err error
		vm, err = m.vmManager.Create(jobName, id, cloudStemcell, deploymentManifest, diskCIDs)
		if err != nil {
			return bosherr.WrapError(err, "Creating VM")
		}

		if err = cloudStemcell.PromoteAsCurrent(); err != nil {
			return bosherr.WrapErrorf(err, "Promoting stemcell as current '%s'", cloudStemcell.CID())
		}

		return nil
	})
	if err != nil {
		return nil, []bidisk.Disk{}, err
	}

	// Create a per-instance blobstore pointing to this VM's agent endpoint.
	blobstore, err := m.blobstoreFactory.Create(vm.MbusURL(), m.blobstoreHTTPClient)
	if err != nil {
		return nil, []bidisk.Disk{}, bosherr.WrapErrorf(err, "Creating blobstore for instance '%s/%d'", jobName, id)
	}

	instance := m.instanceFactory.NewInstance(jobName, id, vm, m.vmManager, m.sshTunnelFactory, blobstore, m.logger)

	if err := instance.WaitUntilReady(eventLoggerStage); err != nil {
		return instance, []bidisk.Disk{}, bosherr.WrapError(err, "Waiting until instance is ready")
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
	skipDrain bool,
	eventLoggerStage biui.Stage,
) error {
	instances, err := m.FindCurrent()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		if err = instance.Delete(pingTimeout, pingDelay, skipDrain, eventLoggerStage); err != nil {
			return bosherr.WrapErrorf(err, "Deleting existing instance '%s/%d'", instance.JobName(), instance.ID())
		}
	}
	return nil
}
