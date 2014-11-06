package deployer

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmins "github.com/cloudfoundry/bosh-micro-cli/deployer/instance"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/deployer/registry"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Deployer interface {
	Deploy(
		bmcloud.Cloud,
		bmdepl.Deployment,
		bmstemcell.ApplySpec,
		bmdepl.Registry,
		bmdepl.SSHTunnel,
		string,
		bmstemcell.CID,
	) error
}

type deployer struct {
	vmManagerFactory   bmvm.ManagerFactory
	diskManagerFactory bmdisk.ManagerFactory
	sshTunnelFactory   bmsshtunnel.Factory
	registryServer     bmregistry.Server
	instanceFactory    bmins.Factory
	eventLoggerStage   bmeventlog.Stage
	logger             boshlog.Logger
	logTag             string
}

func NewDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	diskManagerFactory bmdisk.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	registryServer bmregistry.Server,
	instanceFactory bmins.Factory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("Deploy Micro BOSH", 10)

	return &deployer{
		vmManagerFactory:   vmManagerFactory,
		diskManagerFactory: diskManagerFactory,
		sshTunnelFactory:   sshTunnelFactory,
		registryServer:     registryServer,
		instanceFactory:    instanceFactory,
		eventLoggerStage:   eventLoggerStage,
		logger:             logger,
		logTag:             "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcellApplySpec bmstemcell.ApplySpec,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	mbusURL string,
	stemcellCID bmstemcell.CID,
) error {
	registryReadyErrCh := make(chan error)
	go m.startRegistry(registry, registryReadyErrCh)
	defer m.registryServer.Stop()

	err := <-registryReadyErrCh
	if err != nil {
		return bosherr.WrapError(err, "Starting registry")
	}

	vm, err := m.createVM(cloud, stemcellCID, deployment)
	if err != nil {
		return err
	}

	instance := m.instanceFactory.Create(vm.CID, mbusURL, cloud)

	err = m.waitUntilAgentIsReady(instance, sshTunnelConfig, registry)
	if err != nil {
		return err
	}

	err = m.updateInstance(instance, stemcellApplySpec, deployment)
	if err != nil {
		return err
	}

	err = m.sendStartMessage(instance)
	if err != nil {
		return err
	}

	err = m.waitUntilRunning(instance, deployment.Update.UpdateWatchTime)
	if err != nil {
		return err
	}

	if deploymentJob := deployment.Jobs[0]; deploymentJob.PersistentDisk > 0 {
		err := m.createAndAttachDisk(deploymentJob.PersistentDisk, cloud, vm, instance)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *deployer) startRegistry(registry bmdepl.Registry, readyErrCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyErrCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
}

func (m *deployer) createVM(cloud bmcloud.Cloud, stemcellCID bmstemcell.CID, deployment bmdepl.Deployment) (bmvm.VM, error) {
	vmManager := m.vmManagerFactory.NewManager(cloud)
	eventStep := m.eventLoggerStage.NewStep(fmt.Sprintf("Creating VM from '%s'", stemcellCID))
	eventStep.Start()

	vm, err := vmManager.Create(stemcellCID, deployment)
	if err != nil {
		eventStep.Fail(err.Error())
		return bmvm.VM{}, bosherr.WrapError(err, "Creating VM")
	}
	eventStep.Finish()

	return vm, nil
}

func (m *deployer) waitUntilAgentIsReady(
	instance bmins.Instance,
	sshTunnelConfig bmdepl.SSHTunnel,
	registry bmdepl.Registry,
) error {
	eventStep := m.eventLoggerStage.NewStep("Waiting for the agent")
	eventStep.Start()

	sshTunnelOptions := bmsshtunnel.Options{
		Host:              sshTunnelConfig.Host,
		Port:              sshTunnelConfig.Port,
		User:              sshTunnelConfig.User,
		Password:          sshTunnelConfig.Password,
		PrivateKey:        sshTunnelConfig.PrivateKey,
		LocalForwardPort:  registry.Port,
		RemoteForwardPort: registry.Port,
	}
	sshTunnel := m.sshTunnelFactory.NewSSHTunnel(sshTunnelOptions)
	sshReadyErrCh := make(chan error)
	sshErrCh := make(chan error)
	go sshTunnel.Start(sshReadyErrCh, sshErrCh)
	defer sshTunnel.Stop()

	err := <-sshReadyErrCh
	if err != nil {
		return bosherr.WrapError(err, "Starting SSH tunnel")
	}

	err = instance.WaitToBeReady(300, 500*time.Millisecond)
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Waiting for the instance to be ready")
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) updateInstance(instance bmins.Instance, stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment) error {
	eventStep := m.eventLoggerStage.NewStep("Applying micro BOSH spec")
	eventStep.Start()

	err := instance.Apply(stemcellApplySpec, deployment)
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Updating the instance")
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) sendStartMessage(instance bmins.Instance) error {
	eventStep := m.eventLoggerStage.NewStep("Starting agent services")
	eventStep.Start()

	err := instance.Start()
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Updating the instance")
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) waitUntilRunning(instance bmins.Instance, updateWatchTime bmdepl.WatchTime) error {
	eventStep := m.eventLoggerStage.NewStep("Waiting for the director")
	eventStep.Start()

	time.Sleep(time.Duration(updateWatchTime.Start) * time.Millisecond)
	numAttempts := int((updateWatchTime.End - updateWatchTime.Start) / 1000)

	err := instance.WaitToBeRunning(numAttempts, 1*time.Second)
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Waiting for the director")
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) createAndAttachDisk(diskSize int, cloud bmcloud.Cloud, vm bmvm.VM, instance bmins.Instance) error {
	createEventStep := m.eventLoggerStage.NewStep("Creating disk")
	createEventStep.Start()

	diskManager := m.diskManagerFactory.NewManager(cloud)
	disk, err := diskManager.Create(diskSize, map[string]interface{}{}, vm.CID)
	if err != nil {
		createEventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Creating Disk")
	}
	createEventStep.Finish()

	attachEventStep := m.eventLoggerStage.NewStep("Attaching disk")
	attachEventStep.Start()

	err = instance.AttachDisk(disk)
	if err != nil {
		attachEventStep.Fail(err.Error())
		return err
	}
	attachEventStep.Finish()

	return nil
}
