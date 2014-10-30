package microdeployer

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bminsup "github.com/cloudfoundry/bosh-micro-cli/microdeployer/instanceupdater"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/microdeployer/registry"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/microdeployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
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

type microDeployer struct {
	vmManagerFactory bmvm.ManagerFactory
	sshTunnelFactory bmsshtunnel.Factory
	registryServer   bmregistry.Server
	instanceFactory  bminsup.InstanceFactory
	eventLogger      bmeventlog.EventLogger
	logger           boshlog.Logger
	logTag           string
}

func NewMicroDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	registryServer bmregistry.Server,
	instanceFactory bminsup.InstanceFactory,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *microDeployer {
	return &microDeployer{
		vmManagerFactory: vmManagerFactory,
		sshTunnelFactory: sshTunnelFactory,
		registryServer:   registryServer,
		instanceFactory:  instanceFactory,
		eventLogger:      eventLogger,
		logger:           logger,
		logTag:           "microDeployer",
	}
}

func (m *microDeployer) Deploy(
	cpi bmcloud.Cloud,
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

	vmManager := m.vmManagerFactory.NewManager(cpi)
	_, err = vmManager.CreateVM(stemcellCID, deployment)
	if err != nil {
		return bosherr.WrapError(err, "Creating VM")
	}

	instance := m.instanceFactory.Create(mbusURL)

	err = m.waitUntilAgentIsReady(instance, sshTunnelConfig, registry)
	if err != nil {
		return bosherr.WrapError(err, "Waiting for the agent")
	}

	err = m.updateInstance(instance, stemcellApplySpec, deployment)
	if err != nil {
		return bosherr.WrapError(err, "Updating instance")
	}

	err = m.sendStartMessage(instance)
	if err != nil {
		return bosherr.WrapError(err, "Starting agent services")
	}

	return nil
}

func (m *microDeployer) startRegistry(registry bmdepl.Registry, readyErrCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyErrCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
}

func (m *microDeployer) waitUntilAgentIsReady(
	instance bminsup.Instance,
	sshTunnelConfig bmdepl.SSHTunnel,
	registry bmdepl.Registry,
) error {
	event := bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 4,
		Task:  fmt.Sprintf("Waiting for the agent"),
		Index: 2,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

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
		event = bmeventlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   4,
			Task:    fmt.Sprintf("Waiting for the agent"),
			Index:   2,
			State:   bmeventlog.Failed,
			Message: err.Error(),
		}
		m.eventLogger.AddEvent(event)
		return err
	}

	event = bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 4,
		Task:  fmt.Sprintf("Waiting for the agent"),
		Index: 2,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return nil
}

func (m *microDeployer) updateInstance(instance bminsup.Instance, stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment) error {
	event := bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 4,
		Task:  fmt.Sprintf("Applying micro BOSH spec"),
		Index: 3,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

	err := instance.Apply(stemcellApplySpec, deployment)
	if err != nil {
		event = bmeventlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   4,
			Task:    fmt.Sprintf("Applying micro BOSH spec"),
			Index:   3,
			State:   bmeventlog.Failed,
			Message: err.Error(),
		}
		m.eventLogger.AddEvent(event)

		return bosherr.WrapError(err, "Updating the instance")
	}

	event = bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 4,
		Task:  fmt.Sprintf("Applying micro BOSH spec"),
		Index: 3,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return nil
}

func (m *microDeployer) sendStartMessage(instance bminsup.Instance) error {
	event := bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 4,
		Task:  fmt.Sprintf("Starting agent services"),
		Index: 4,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

	err := instance.Start()
	if err != nil {
		event = bmeventlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   4,
			Task:    fmt.Sprintf("Starting agent services"),
			Index:   4,
			State:   bmeventlog.Failed,
			Message: err.Error(),
		}
		m.eventLogger.AddEvent(event)

		return bosherr.WrapError(err, "Updating the instance")
	}

	event = bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 4,
		Task:  fmt.Sprintf("Starting agent services"),
		Index: 4,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return nil
}
