package vm

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Deployer interface {
	Deploy(
		cloud bmcloud.Cloud,
		deployment bmdepl.Deployment,
		stemcell bmstemcell.CloudStemcell,
		sshTunnelOptions bmsshtunnel.Options,
		mbusURL string,
		eventLoggerStage bmeventlog.Stage,
	) (VM, error)
}

type vmDeployer struct {
	vmManagerFactory ManagerFactory
	sshTunnelFactory bmsshtunnel.Factory
}

func NewDeployer(vmManagerFactory ManagerFactory, sshTunnelFactory bmsshtunnel.Factory) Deployer {
	return &vmDeployer{
		vmManagerFactory: vmManagerFactory,
		sshTunnelFactory: sshTunnelFactory,
	}
}

func (d *vmDeployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcell bmstemcell.CloudStemcell,
	sshTunnelOptions bmsshtunnel.Options,
	mbusURL string,
	eventLoggerStage bmeventlog.Stage,
) (VM, error) {
	eventStep := eventLoggerStage.NewStep(fmt.Sprintf("Creating VM from stemcell '%s'", stemcell.CID))
	eventStep.Start()

	vm, err := d.createVM(cloud, deployment, stemcell, mbusURL)
	if err != nil {
		eventStep.Fail(err.Error())
		return nil, err
	}

	eventStep.Finish()

	eventStep = eventLoggerStage.NewStep(fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID()))
	eventStep.Start()

	err = d.waitUntilAgentIsReady(vm, sshTunnelOptions)
	if err != nil {
		eventStep.Fail(err.Error())
		return nil, err
	}

	eventStep.Finish()

	return vm, nil
}

func (d *vmDeployer) createVM(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcell bmstemcell.CloudStemcell,
	mbusURL string,
) (VM, error) {
	vmManager := d.vmManagerFactory.NewManager(cloud)
	vm, err := vmManager.Create(stemcell, deployment, mbusURL)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating VM")
	}

	return vm, nil
}

func (d *vmDeployer) waitUntilAgentIsReady(
	vm VM,
	sshTunnelOptions bmsshtunnel.Options,
) error {
	sshTunnel := d.sshTunnelFactory.NewSSHTunnel(sshTunnelOptions)
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
		return bosherr.WrapError(err, "Waiting for the vm to be ready")
	}

	return nil
}
