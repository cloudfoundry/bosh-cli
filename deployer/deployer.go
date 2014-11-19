package deployer

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
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
		bmstemcell.CloudStemcell,
	) error
}

type deployer struct {
	vmDeployer       VMDeployer
	diskDeployer     DiskDeployer
	registryServer   bmregistry.Server
	eventLoggerStage bmeventlog.Stage
	logger           boshlog.Logger
	logTag           string
}

func NewDeployer(
	vmDeployer VMDeployer,
	diskDeployer DiskDeployer,
	registryServer bmregistry.Server,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *deployer {
	eventLoggerStage := eventLogger.NewStage("deploying")

	return &deployer{
		vmDeployer:       vmDeployer,
		diskDeployer:     diskDeployer,
		registryServer:   registryServer,
		eventLoggerStage: eventLoggerStage,
		logger:           logger,
		logTag:           "deployer",
	}
}

func (m *deployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcellApplySpec bmstemcell.ApplySpec,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	mbusURL string,
	stemcell bmstemcell.CloudStemcell,
) error {
	m.eventLoggerStage.Start()

	registryReadyErrCh := make(chan error)
	go m.startRegistry(registry, registryReadyErrCh)
	defer m.registryServer.Stop()

	err := <-registryReadyErrCh
	if err != nil {
		return bosherr.WrapError(err, "Starting registry")
	}

	sshTunnelOptions := bmsshtunnel.Options{
		Host:              sshTunnelConfig.Host,
		Port:              sshTunnelConfig.Port,
		User:              sshTunnelConfig.User,
		Password:          sshTunnelConfig.Password,
		PrivateKey:        sshTunnelConfig.PrivateKey,
		LocalForwardPort:  registry.Port,
		RemoteForwardPort: registry.Port,
	}

	vm, err := m.vmDeployer.Deploy(cloud, deployment, stemcell, sshTunnelOptions, mbusURL, m.eventLoggerStage)
	if err != nil {
		return bosherr.WrapError(err, "Deploying VM")
	}

	jobName := deployment.Jobs[0].Name

	diskPool, err := deployment.DiskPool(jobName)
	if err != nil {
		return bosherr.WrapError(err, "Getting disk pool")
	}

	err = m.diskDeployer.Deploy(diskPool, cloud, vm, m.eventLoggerStage)
	if err != nil {
		return bosherr.WrapError(err, "Deploying disk")
	}

	err = m.startVM(vm, stemcellApplySpec, deployment, jobName)
	if err != nil {
		return err
	}

	err = m.waitUntilRunning(vm, deployment.Update.UpdateWatchTime, jobName)
	if err != nil {
		return err
	}

	m.eventLoggerStage.Finish()
	return nil
}

func (m *deployer) startRegistry(registry bmdepl.Registry, readyErrCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyErrCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
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
