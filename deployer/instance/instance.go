package instance

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployer/sshtunnel"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
)

type Instance interface {
	JobName() string
	ID() int
	WaitUntilReady(bmdepl.Registry, bmdepl.SSHTunnel, bmeventlog.Stage) error
	StartJobs(newState bmstemcell.ApplySpec, deploymentManifest bmdepl.Manifest, eventLoggerStage bmeventlog.Stage) error
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
	registryConfig bmdepl.Registry,
	sshTunnelConfig bmdepl.SSHTunnel,
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

// StartJobs sends the agent a new apply spec, restarts the agent, and polls until the agent says the jobs are running
func (i *instance) StartJobs(newState bmstemcell.ApplySpec, deploymentManifest bmdepl.Manifest, eventLoggerStage bmeventlog.Stage) error {
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
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", i.vm.CID())
	waitingForAgentErr := eventLoggerStage.PerformStep(stepName, func() error {
		//TODO: do we need to start an ssh tunnel so that the vm can read from the registry?
		if err := i.vm.WaitUntilReady(pingTimeout, pingDelay); err != nil {
			return bosherr.WrapError(err, "Agent unreachable")
		}
		return nil
	})
	// if agent responds, delete vm, otherwise continue
	if waitingForAgentErr != nil {
		i.logger.Warn(i.logTag, "Gave up waiting for agent: %s", waitingForAgentErr.Error())
	} else {
		if err := i.stopJobs(eventLoggerStage); err != nil {
			return err
		}
		if err := i.unmountDisks(eventLoggerStage); err != nil {
			return err
		}
	}

	stepName = fmt.Sprintf("Deleting VM '%s'", i.vm.CID())
	return eventLoggerStage.PerformStep(stepName, func() error {
		if err := i.vm.Delete(); err != nil {
			return bosherr.WrapError(err, "Deleting VM")
		}
		return nil
	})
}

func (i *instance) startJobs(
	vm bmvm.VM,
	stemcellApplySpec bmstemcell.ApplySpec,
	deploymentManifest bmdepl.Manifest,
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

func (i *instance) waitUntilJobsAreRunning(deploymentManifest bmdepl.Manifest, eventLoggerStage bmeventlog.Stage) error {
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
