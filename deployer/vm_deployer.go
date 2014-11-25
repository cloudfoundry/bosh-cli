package deployer

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type VMDeployer interface {
	Deploy(
		cloud bmcloud.Cloud,
		deployment bmdepl.Deployment,
		stemcell bmstemcell.CloudStemcell,
		mbusURL string,
		eventLoggerStage bmeventlog.Stage,
	) (bmvm.VM, error)

	WaitUntilReady(
		vm bmvm.VM,
		sshTunnelOptions bmsshtunnel.Options,
		eventLoggerStage bmeventlog.Stage,
	) error
}

type vmDeployer struct {
	vmManagerFactory bmvm.ManagerFactory
	sshTunnelFactory bmsshtunnel.Factory
	logger           boshlog.Logger
	logTag           string
}

func NewVMDeployer(
	vmManagerFactory bmvm.ManagerFactory,
	sshTunnelFactory bmsshtunnel.Factory,
	logger boshlog.Logger,
) VMDeployer {
	return &vmDeployer{
		vmManagerFactory: vmManagerFactory,
		sshTunnelFactory: sshTunnelFactory,
		logger:           logger,
		logTag:           "vmDeployer",
	}
}

func (d *vmDeployer) Deploy(
	cloud bmcloud.Cloud,
	deployment bmdepl.Deployment,
	stemcell bmstemcell.CloudStemcell,
	mbusURL string,
	eventLoggerStage bmeventlog.Stage,
) (bmvm.VM, error) {
	vmManager := d.vmManagerFactory.NewManager(cloud, mbusURL)

	jobName := deployment.Jobs[0].Name
	err := d.deleteExistingVM(vmManager, eventLoggerStage, jobName)
	if err != nil {
		return nil, err
	}

	eventStep := eventLoggerStage.NewStep(fmt.Sprintf("Creating VM from stemcell '%s'", stemcell.CID()))
	eventStep.Start()

	vm, err := vmManager.Create(stemcell, deployment)
	if err != nil {
		err = bosherr.WrapError(err, "Creating VM")
		eventStep.Fail(err.Error())
		return nil, err
	}

	err = stemcell.PromoteAsCurrent()
	if err != nil {
		return nil, bosherr.WrapError(err, "Promoting stemcell as current '%s'", stemcell.CID())
	}

	eventStep.Finish()

	return vm, nil
}

func (d *vmDeployer) WaitUntilReady(vm bmvm.VM, sshTunnelOptions bmsshtunnel.Options, eventLoggerStage bmeventlog.Stage) error {
	eventStep := eventLoggerStage.NewStep(fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID()))
	eventStep.Start()

	if !sshTunnelOptions.IsEmpty() {
		sshTunnel := d.sshTunnelFactory.NewSSHTunnel(sshTunnelOptions)
		sshReadyErrCh := make(chan error)
		sshErrCh := make(chan error)
		go sshTunnel.Start(sshReadyErrCh, sshErrCh)
		defer sshTunnel.Stop()

		err := <-sshReadyErrCh
		if err != nil {
			return bosherr.WrapError(err, "Starting SSH tunnel")
		}
	}

	err := vm.WaitToBeReady(10*time.Minute, 500*time.Millisecond)
	if err != nil {
		err = bosherr.WrapError(err, "Waiting for the vm to be ready")
		eventStep.Fail(err.Error())
		return err
	}

	eventStep.Finish()

	return nil
}

func (d *vmDeployer) deleteExistingVM(vmManager bmvm.Manager, eventLoggerStage bmeventlog.Stage, jobName string) error {
	vm, found, err := vmManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding existing VM")
	}

	if found {
		waitingForAgentStep := eventLoggerStage.NewStep(fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID()))
		waitingForAgentStep.Start()

		err = vm.WaitToBeReady(10*time.Second, 500*time.Millisecond)
		if err != nil {
			err = bosherr.WrapError(err, "Agent unreachable")
			waitingForAgentStep.Fail(err.Error())
		} else {
			waitingForAgentStep.Finish()

			stopVMStep := eventLoggerStage.NewStep(fmt.Sprintf("Stopping '%s'", jobName))
			stopVMStep.Start()
			err = vm.Stop()
			if err != nil {
				err = bosherr.WrapError(err, "Stopping VM")
				stopVMStep.Fail(err.Error())
				return err
			}
			stopVMStep.Finish()

			disks, err := vm.Disks()
			if err != nil {
				return bosherr.WrapError(err, "Getting VM '%s' disks", vm.CID())
			}

			for _, disk := range disks {
				unmountDiskStep := eventLoggerStage.NewStep(fmt.Sprintf("Unmounting disk '%s'", disk.CID()))
				unmountDiskStep.Start()
				err = vm.UnmountDisk(disk)
				if err != nil {
					err = bosherr.WrapError(err, "Unmounting disk '%s' from VM '%s'", disk.CID(), vm.CID())
					unmountDiskStep.Fail(err.Error())
					return err
				}
				unmountDiskStep.Finish()
			}
		}

		deleteVMStep := eventLoggerStage.NewStep(fmt.Sprintf("Deleting VM '%s'", vm.CID()))
		deleteVMStep.Start()

		err = vm.Delete()
		if err != nil {
			err = bosherr.WrapError(err, "Deleting VM")
			deleteVMStep.Fail(err.Error())
			return err
		}

		deleteVMStep.Finish()
	}

	return nil
}
