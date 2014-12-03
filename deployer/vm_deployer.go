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
) (vm bmvm.VM, err error) {
	vmManager := d.vmManagerFactory.NewManager(cloud, mbusURL)

	jobName := deployment.Jobs[0].Name
	err = d.deleteExistingVM(vmManager, eventLoggerStage, jobName)
	if err != nil {
		return nil, err
	}

	stepName := fmt.Sprintf("Creating VM from stemcell '%s'", stemcell.CID())
	err = eventLoggerStage.PerformStep(stepName, func() error {
		vm, err = vmManager.Create(stemcell, deployment)
		if err != nil {
			return bosherr.WrapError(err, "Creating VM")
		}

		if err = stemcell.PromoteAsCurrent(); err != nil {
			return bosherr.WrapErrorf(err, "Promoting stemcell as current '%s'", stemcell.CID())
		}

		return nil
	})

	return vm, err
}

func (d *vmDeployer) WaitUntilReady(vm bmvm.VM, sshTunnelOptions bmsshtunnel.Options, eventLoggerStage bmeventlog.Stage) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID())
	err := eventLoggerStage.PerformStep(stepName, func() error {
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
			return bosherr.WrapError(err, "Waiting for the vm to be ready")
		}

		return nil
	})

	return err
}

func (d *vmDeployer) deleteExistingVM(vmManager bmvm.Manager, eventLoggerStage bmeventlog.Stage, jobName string) error {
	vm, found, err := vmManager.FindCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Finding existing VM")
	}

	if found {
		stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", vm.CID())
		waitingForAgentErr := eventLoggerStage.PerformStep(stepName, func() error {
			if err = vm.WaitToBeReady(10*time.Second, 500*time.Millisecond); err != nil {
				return bosherr.WrapError(err, "Agent unreachable")
			}
			return nil
		})
		// if agent responds, delete vm, otherwise continue
		if waitingForAgentErr == nil {
			if err = d.shutDownJob(jobName, vm, eventLoggerStage); err != nil {
				return err
			}
		}

		stepName = fmt.Sprintf("Deleting VM '%s'", vm.CID())
		err = eventLoggerStage.PerformStep(stepName, func() error {
			if err = vm.Delete(); err != nil {
				return bosherr.WrapError(err, "Deleting VM")
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *vmDeployer) shutDownJob(jobName string, vm bmvm.VM, eventLoggerStage bmeventlog.Stage) error {
	stepNAme := fmt.Sprintf("Stopping job '%s'", jobName)
	err := eventLoggerStage.PerformStep(stepNAme, func() error {
		err := vm.Stop()
		if err != nil {
			return bosherr.WrapErrorf(err, "Stopping job '%s'", jobName)
		}
		return nil
	})
	if err != nil {
		return err
	}

	disks, err := vm.Disks()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting VM '%s' disks", vm.CID())
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Unmounting disk '%s'", disk.CID())
		err = eventLoggerStage.PerformStep(stepName, func() error {
			if err = vm.UnmountDisk(disk); err != nil {
				return bosherr.WrapErrorf(err, "Unmounting disk '%s' from VM '%s'", disk.CID(), vm.CID())
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
