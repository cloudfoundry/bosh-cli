package microdeployer

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogging"
	bmregistry "github.com/cloudfoundry/bosh-micro-cli/registry"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/retrystrategy"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"
)

type Deployer interface {
	Deploy(bmcloud.Cloud, bmdepl.Deployment, bmdepl.Registry, bmdepl.SSHTunnel, bmretrystrategy.RetryStrategy, bmstemcell.CID) error
}

type microDeployer struct {
	vmManagerFactory bmvm.ManagerFactory
	sshTunnelFactory bmsshtunnel.Factory
	registryServer   bmregistry.Server
	eventLogger      bmeventlog.EventLogger
	logger           boshlog.Logger
	logTag           string
}

func NewMicroDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	registryServer bmregistry.Server,
	eventLogger bmeventlog.EventLogger,
	logger boshlog.Logger,
) *microDeployer {
	return &microDeployer{
		vmManagerFactory: vmManagerFactory,
		sshTunnelFactory: sshTunnelFactory,
		registryServer:   registryServer,
		eventLogger:      eventLogger,
		logger:           logger,
		logTag:           "microDeployer",
	}
}

func (m *microDeployer) Deploy(
	cpi bmcloud.Cloud,
	deployment bmdepl.Deployment,
	registry bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
	agentPingRetryStrategy bmretrystrategy.RetryStrategy,
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

	err = <-sshReadyErrCh
	if err != nil {
		return bosherr.WrapError(err, "Starting SSH tunnel")
	}

	err = m.waitUntilAgentIsReady(agentPingRetryStrategy)
	if err != nil {
		return bosherr.WrapError(err, "Waiting for the agent")
	}

	return nil
}

func (m *microDeployer) startRegistry(registry bmdepl.Registry, readyErrCh chan error) {
	err := m.registryServer.Start(registry.Username, registry.Password, registry.Host, registry.Port, readyErrCh)
	if err != nil {
		m.logger.Debug(m.logTag, "Registry error occurred: %s", err.Error())
	}
}

func (m *microDeployer) waitUntilAgentIsReady(agentPingRetryStrategy bmretrystrategy.RetryStrategy) error {
	event := bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
		Task:  fmt.Sprintf("Waiting for the agent"),
		Index: 1,
		State: bmeventlog.Started,
	}
	m.eventLogger.AddEvent(event)

	err := agentPingRetryStrategy.Try()
	if err != nil {
		event = bmeventlog.Event{
			Stage:   "Deploy Micro BOSH",
			Total:   1,
			Task:    fmt.Sprintf("Waiting for the agent"),
			Index:   1,
			State:   bmeventlog.Failed,
			Message: err.Error(),
		}
		m.eventLogger.AddEvent(event)
		return err
	}

	event = bmeventlog.Event{
		Stage: "Deploy Micro BOSH",
		Total: 1,
		Task:  fmt.Sprintf("Waiting for the agent"),
		Index: 1,
		State: bmeventlog.Finished,
	}
	m.eventLogger.AddEvent(event)

	return nil
}
