package deployer

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
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
	eventLoggerStage   bmeventlog.Stage
	logger             boshlog.Logger
	logTag             string
}

func NewDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	diskManagerFactory bmdisk.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	registryServer bmregistry.Server,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		vmManagerFactory:   vmManagerFactory,
		diskManagerFactory: diskManagerFactory,
		sshTunnelFactory:   sshTunnelFactory,
		registryServer:     registryServer,
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
	m.eventLoggerStage.Start()
	defer m.eventLoggerStage.Finish()

	registryReadyErrCh := make(chan error)
	go m.startRegistry(registry, registryReadyErrCh)
	defer m.registryServer.Stop()

	err := <-registryReadyErrCh
	if err != nil {
		return bosherr.WrapError(err, "Starting registry")
	}

	vm, err := m.createVM(cloud, stemcellCID, deployment, mbusURL)
	if err != nil {
		return err
	}

	err = m.waitUntilAgentIsReady(vm, sshTunnelConfig, registry)
	if err != nil {
		return err
	}

	if deploymentJob := deployment.Jobs[0]; deploymentJob.PersistentDisk > 0 {
		err := m.createAndAttachDisk(deploymentJob.PersistentDisk, cloud, vm)
		if err != nil {
			return err
		}
	}

	jobName := deployment.Jobs[0].Name
	err = m.startVM(vm, stemcellApplySpec, deployment, jobName)
	if err != nil {
		return err
	}

	err = m.waitUntilRunning(vm, deployment.Update.UpdateWatchTime, jobName)
	if err != nil {
		return err
	}

	return nil
}

func (m *deployer) startRegistry(registry bmdepl.Registry, readyErrCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyErrCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
}

func (m *deployer) createVM(
	cloud bmcloud.Cloud,
	stemcellCID bmstemcell.CID,
	deployment bmdepl.Deployment,
	mbusURL string,
) (bmvm.VM, error) {
	vmManager := m.vmManagerFactory.NewManager(cloud)
	eventStep := m.eventLoggerStage.NewStep(fmt.Sprintf("Creating VM from stemcell '%s'", stemcellCID))
	eventStep.Start()

	vm, err := vmManager.Create(stemcellCID, deployment, mbusURL)
	if err != nil {
		eventStep.Fail(err.Error())
		return nil, bosherr.WrapError(err, "Creating VM")
	}
	eventStep.Finish()

	return vm, nil
}

func (m *deployer) waitUntilAgentIsReady(
	vm bmvm.VM,
	sshTunnelConfig bmdepl.SSHTunnel,
	registry bmdepl.Registry,
) error {
	eventStep := m.eventLoggerStage.NewStep(fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID()))
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

	err = vm.WaitToBeReady(300, 500*time.Millisecond)
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Waiting for the vm to be ready")
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) startVM(vm bmvm.VM, stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment, jobName string) error {
	eventStep := m.eventLoggerStage.NewStep(fmt.Sprintf("Starting '%s'", jobName))
	eventStep.Start()

	err := vm.Apply(stemcellApplySpec, deployment)
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Updating the vm")
	}

	err = vm.Start()
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Starting vm")
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) waitUntilRunning(vm bmvm.VM, updateWatchTime bmdepl.WatchTime, jobName string) error {
	eventStep := m.eventLoggerStage.NewStep(fmt.Sprintf("Waiting for '%s'", jobName))
	eventStep.Start()

	time.Sleep(time.Duration(updateWatchTime.Start) * time.Millisecond)
	numAttempts := int((updateWatchTime.End - updateWatchTime.Start) / 1000)

	err := vm.WaitToBeRunning(numAttempts, 1*time.Second)
	if err != nil {
		eventStep.Fail(err.Error())
		return bosherr.WrapError(err, fmt.Sprintf("Waiting for '%s'", jobName))
	}

	eventStep.Finish()

	return nil
}

func (m *deployer) createAndAttachDisk(diskSize int, cloud bmcloud.Cloud, vm bmvm.VM) error {
	createEventStep := m.eventLoggerStage.NewStep("Creating disk")
	createEventStep.Start()

	diskManager := m.diskManagerFactory.NewManager(cloud)
	disk, err := diskManager.Create(diskSize, map[string]interface{}{}, vm.CID())
	if err != nil {
		createEventStep.Fail(err.Error())
		return bosherr.WrapError(err, "Creating Disk")
	}
	createEventStep.Finish()

	attachEventStep := m.eventLoggerStage.NewStep(fmt.Sprintf("Attaching disk '%s' to VM '%s'", disk.CID(), vm.CID()))
	attachEventStep.Start()

	err = vm.AttachDisk(disk)
	if err != nil {
		attachEventStep.Fail(err.Error())
		return err
	}
	attachEventStep.Finish()

	return nil
}
