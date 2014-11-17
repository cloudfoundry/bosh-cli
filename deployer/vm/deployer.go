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

func NewDeployer(
	vmManagerFactory ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
) Deployer {
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
	vmManager := d.vmManagerFactory.NewManager(cloud, mbusURL)

	vm, found, err := vmManager.FindCurrent()
	if err != nil {
		return vm, bosherr.WrapError(err, "Finding existing VM")
	}

	if found {
		eventStep := eventLoggerStage.NewStep(fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID()))
		eventStep.Start()

		err = vm.WaitToBeReady(10*time.Second, 500*time.Millisecond)
		if err != nil {
			err = bosherr.WrapError(err, "Pinging existing VM")
			eventStep.Fail(err.Error())
			return vm, err
		}

		eventStep.Finish()
	}

	eventStep := eventLoggerStage.NewStep(fmt.Sprintf("Creating VM from stemcell '%s'", stemcell.CID))
	eventStep.Start()

	vm, err = vmManager.Create(stemcell, deployment)
	if err != nil {
		err = bosherr.WrapError(err, "Creating VM")
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

	err = vm.WaitToBeReady(10*time.Minute, 500*time.Millisecond)
	if err != nil {
		return bosherr.WrapError(err, "Waiting for the vm to be ready")
	}

	return nil
}
