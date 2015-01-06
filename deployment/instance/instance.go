package instance

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployment/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

type Instance interface {
	JobName() string
	ID() int
	WaitUntilReady(bminstallmanifest.Registry, bminstallmanifest.SSHTunnel, bmeventlog.Stage) error
	UpdateDisks(bmdeplmanifest.Manifest, bmeventlog.Stage) error
	StartJobs(newState bmstemcell.ApplySpec, deploymentManifest bmdeplmanifest.Manifest, eventLoggerStage bmeventlog.Stage) error
	Delete(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		eventLoggerStage bmeventlog.Stage,
	) error
}

type instance struct {
	jobName          string
	id               int
	vm               bmvm.VM
	vmManager        bmvm.Manager
	sshTunnelFactory bmsshtunnel.Factory
	logger           boshlog.Logger
	logTag           string
}

func NewInstance(
	jobName string,
	id int,
	vm bmvm.VM,
	vmManager bmvm.Manager,
	sshTunnelFactory bmsshtunnel.Factory,
	logger boshlog.Logger,
) Instance {
	return &instance{
		jobName:          jobName,
		id:               id,
		vm:               vm,
		vmManager:        vmManager,
		sshTunnelFactory: sshTunnelFactory,
		logger:           logger,
		logTag:           "instance",
	}
}

func (i *instance) JobName() string {
	return i.jobName
}

func (i *instance) ID() int {
	return i.id
}

func (i *instance) WaitUntilReady(
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	eventLoggerStage bmeventlog.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s' to be ready", i.vm.CID())
	err := eventLoggerStage.PerformStep(stepName, func() error {
		if !registryConfig.IsEmpty() && !sshTunnelConfig.IsEmpty() {
			sshTunnelOptions := bmsshtunnel.Options{
				Host:              sshTunnelConfig.Host,
				Port:              sshTunnelConfig.Port,
				User:              sshTunnelConfig.User,
				Password:          sshTunnelConfig.Password,
				PrivateKey:        sshTunnelConfig.PrivateKey,
				LocalForwardPort:  registryConfig.Port,
				RemoteForwardPort: registryConfig.Port,
			}
			sshTunnel := i.sshTunnelFactory.NewSSHTunnel(sshTunnelOptions)
			sshReadyErrCh := make(chan error)
			sshErrCh := make(chan error)
			go sshTunnel.Start(sshReadyErrCh, sshErrCh)
			defer sshTunnel.Stop()

			err := <-sshReadyErrCh
			if err != nil {
				return bosherr.WrapError(err, "Starting SSH tunnel")
			}
		}

		return i.vm.WaitUntilReady(10*time.Minute, 500*time.Millisecond)
	})

	return err
}

func (i *instance) UpdateDisks(deploymentManifest bmdeplmanifest.Manifest, eventLoggerStage bmeventlog.Stage) error {
	diskPool, err := deploymentManifest.DiskPool(i.jobName)
	if err != nil {
		return bosherr.WrapError(err, "Getting disk pool")
	}

	err = i.vm.UpdateDisks(diskPool, eventLoggerStage)
	if err != nil {
		return bosherr.WrapError(err, "Updating disks")
	}

	return nil
}

// StartJobs sends the agent a new apply spec, restarts the agent, and polls until the agent says the jobs are running
func (i *instance) StartJobs(newState bmstemcell.ApplySpec, deploymentManifest bmdeplmanifest.Manifest, eventLoggerStage bmeventlog.Stage) error {
	if err := i.startJobs(i.vm, newState, deploymentManifest, eventLoggerStage); err != nil {
		return err
	}

	return i.waitUntilJobsAreRunning(deploymentManifest, eventLoggerStage)
}

func (i *instance) Delete(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	eventLoggerStage bmeventlog.Stage,
) error {
	vmExists, err := i.vm.Exists()
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking existance of vm for instance '%s/%d'", i.jobName, i.id)
	}

	if vmExists {
		if err = i.shutdown(pingTimeout, pingDelay, eventLoggerStage); err != nil {
			return err
		}
	}

	// non-existent VMs still need to be 'deleted' to clean up related resources owned by the CPI
	stepName := fmt.Sprintf("Deleting VM '%s'", i.vm.CID())
	return eventLoggerStage.PerformStep(stepName, func() error {
		err := i.vm.Delete()
		cloudErr, ok := err.(bmcloud.Error)
		if ok && cloudErr.Type() == bmcloud.VMNotFoundError {
			return bmeventlog.NewSkippedStepError(cloudErr.Error())
		}
		return err
	})
}

func (i *instance) shutdown(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	eventLoggerStage bmeventlog.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", i.vm.CID())
	waitingForAgentErr := eventLoggerStage.PerformStep(stepName, func() error {
		if err := i.vm.WaitUntilReady(pingTimeout, pingDelay); err != nil {
			return bosherr.WrapError(err, "Agent unreachable")
		}
		return nil
	})
	if waitingForAgentErr != nil {
		i.logger.Warn(i.logTag, "Gave up waiting for agent: %s", waitingForAgentErr.Error())
		return nil
	}

	if err := i.stopJobs(eventLoggerStage); err != nil {
		return err
	}
	if err := i.unmountDisks(eventLoggerStage); err != nil {
		return err
	}
	return nil
}

func (i *instance) startJobs(
	vm bmvm.VM,
	stemcellApplySpec bmstemcell.ApplySpec,
	deploymentManifest bmdeplmanifest.Manifest,
	eventLoggerStage bmeventlog.Stage,
) error {
	stepName := fmt.Sprintf("Starting instance '%s/%d'", i.jobName, i.id)
	return eventLoggerStage.PerformStep(stepName, func() error {
		err := vm.Apply(stemcellApplySpec, deploymentManifest)
		if err != nil {
			return bosherr.WrapError(err, "Applying the agent state")
		}

		//TODO: move this 'Start' in here, because it's telling the agent to start, not the vm...
		err = vm.Start()
		if err != nil {
			return bosherr.WrapError(err, "Starting the agent")
		}

		return nil
	})
}

func (i *instance) waitUntilJobsAreRunning(deploymentManifest bmdeplmanifest.Manifest, eventLoggerStage bmeventlog.Stage) error {
	updateWatchTime := deploymentManifest.Update.UpdateWatchTime
	start := time.Duration(updateWatchTime.Start) * time.Millisecond
	end := time.Duration(updateWatchTime.End) * time.Millisecond
	delayBetweenAttempts := 1 * time.Second
	maxAttempts := int((end - start) / delayBetweenAttempts)

	stepName := fmt.Sprintf("Waiting for instance '%s/%d' to be running", i.jobName, i.id)
	return eventLoggerStage.PerformStep(stepName, func() error {
		time.Sleep(start)
		return i.vm.WaitToBeRunning(maxAttempts, delayBetweenAttempts)
	})
}

func (i *instance) stopJobs(eventLoggerStage bmeventlog.Stage) error {
	stepName := fmt.Sprintf("Stopping jobs on instance '%s/%d'", i.jobName, i.id)
	return eventLoggerStage.PerformStep(stepName, func() error {
		return i.vm.Stop()
	})
}

func (i *instance) unmountDisks(eventLoggerStage bmeventlog.Stage) error {
	disks, err := i.vm.Disks()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting VM '%s' disks", i.vm.CID())
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Unmounting disk '%s'", disk.CID())
		err = eventLoggerStage.PerformStep(stepName, func() error {
			if err := i.vm.UnmountDisk(disk); err != nil {
				return bosherr.WrapErrorf(err, "Unmounting disk '%s' from VM '%s'", disk.CID(), i.vm.CID())
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
