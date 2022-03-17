package instance

import (
	"fmt"
	"time"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	biinstancestate "github.com/cloudfoundry/bosh-cli/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Instance interface {
	JobName() string
	ID() int
	Disks() ([]bidisk.Disk, error)
	WaitUntilReady(biui.Stage) error
	UpdateDisks(bideplmanifest.Manifest, biui.Stage) ([]bidisk.Disk, error)
	UpdateJobs(bideplmanifest.Manifest, biui.Stage) error
	Delete(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		skipDrain bool,
		stage biui.Stage,
	) error
	Stop(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		skipDrain bool,
		stage biui.Stage,
	) error
	Start(
		update bideplmanifest.Update,
		pingTimeout time.Duration,
		pingDelay time.Duration,
		stage biui.Stage,
	) error
}

type instance struct {
	jobName          string
	id               int
	vm               bivm.VM
	vmManager        bivm.Manager
	sshTunnelFactory bisshtunnel.Factory
	stateBuilder     biinstancestate.Builder
	logger           boshlog.Logger
	logTag           string
}

func NewInstance(
	jobName string,
	id int,
	vm bivm.VM,
	vmManager bivm.Manager,
	sshTunnelFactory bisshtunnel.Factory,
	stateBuilder biinstancestate.Builder,
	logger boshlog.Logger,
) Instance {
	return &instance{
		jobName:          jobName,
		id:               id,
		vm:               vm,
		vmManager:        vmManager,
		sshTunnelFactory: sshTunnelFactory,
		stateBuilder:     stateBuilder,
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

func (i *instance) Disks() ([]bidisk.Disk, error) {
	disks, err := i.vm.Disks()
	if err != nil {
		return disks, bosherr.WrapError(err, "Listing instance disks")
	}
	return disks, nil
}

func (i *instance) WaitUntilReady(
	stage biui.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s' to be ready", i.vm.CID())
	err := stage.Perform(stepName, func() error {

		return i.vm.WaitUntilReady(10*time.Minute, 500*time.Millisecond)
	})

	return err
}

func (i *instance) UpdateDisks(deploymentManifest bideplmanifest.Manifest, stage biui.Stage) ([]bidisk.Disk, error) {
	diskPool, err := deploymentManifest.DiskPool(i.jobName)
	if err != nil {
		return []bidisk.Disk{}, bosherr.WrapError(err, "Getting disk pool")
	}

	disks, err := i.vm.UpdateDisks(diskPool, stage)
	if err != nil {
		return disks, bosherr.WrapError(err, "Updating disks")
	}

	return disks, nil
}

func (i *instance) UpdateJobs(
	deploymentManifest bideplmanifest.Manifest,
	stage biui.Stage,
) error {
	initialAgentState, err := i.stateBuilder.BuildInitialState(i.jobName, i.id, deploymentManifest)
	if err != nil {
		return bosherr.WrapErrorf(err, "Building initial state for instance '%s/%d'", i.jobName, i.id)
	}

	// apply it to agent to force it to load networking details
	err = i.vm.Apply(initialAgentState.ToApplySpec())
	if err != nil {
		return bosherr.WrapError(err, "Applying the initial agent state")
	}

	// now that the agent will tell us the address, get new state
	resolvedAgentState, err := i.vm.GetState()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting state for instance '%s/%d'", i.jobName, i.id)
	}

	newAgentState, err := i.stateBuilder.Build(i.jobName, i.id, deploymentManifest, stage, resolvedAgentState)
	if err != nil {
		return bosherr.WrapErrorf(err, "Building state for instance '%s/%d'", i.jobName, i.id)
	}
	stepName := fmt.Sprintf("Updating instance '%s/%d'", i.jobName, i.id)
	err = stage.Perform(stepName, func() error {
		err = i.vm.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping the agent")
		}

		err = i.vm.Apply(newAgentState.ToApplySpec())
		if err != nil {
			return bosherr.WrapError(err, "Applying the agent state")
		}

		err = i.vm.RunScript("pre-start", map[string]interface{}{})
		if err != nil {
			return bosherr.WrapError(err, "Running the pre-start script")
		}

		err = i.vm.Start()
		if err != nil {
			return bosherr.WrapError(err, "Starting the agent")
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = i.waitUntilJobsAreRunning(deploymentManifest.Update.UpdateWatchTime, stage)
	if err != nil {
		return err
	}

	err = i.runScriptInStep(stage, "post-start", i.jobName, i.id)

	return err
}

func (i *instance) Delete(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	skipDrain bool,
	stage biui.Stage,
) error {
	vmExists, err := i.vm.Exists()
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking existence of vm for instance '%s/%d'", i.jobName, i.id)
	}

	if vmExists {
		if err = i.shutdown(pingTimeout, pingDelay, skipDrain, false, stage); err != nil {
			return err
		}
	}

	// non-existent VMs still need to be 'deleted' to clean up related resources owned by the CPI
	stepName := fmt.Sprintf("Deleting VM '%s'", i.vm.CID())
	return stage.Perform(stepName, func() error {
		err := i.vm.Delete()
		cloudErr, ok := err.(bicloud.Error)
		if ok && cloudErr.Type() == bicloud.VMNotFoundError {
			return biui.NewSkipStageError(cloudErr, "VM not found")
		}
		return err
	})
}

func (i *instance) Stop(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	skipDrain bool,
	stage biui.Stage,
) error {

	return i.shutdown(pingTimeout, pingDelay, skipDrain, true, stage) //; err != nil {
}

func (i *instance) Start(
	update bideplmanifest.Update,
	pingTimeout time.Duration,
	pingDelay time.Duration,
	stage biui.Stage,
) error {

	if waitingForAgentErr := i.waitForAgent(pingDelay, pingTimeout, stage); waitingForAgentErr != nil {
		i.logger.Warn(i.logTag, "Gave up waiting for agent: %s", waitingForAgentErr.Error())
		return nil
	}

	if err := i.runScriptInStep(stage, "pre-start", i.jobName, i.id); err != nil {
		return err
	}

	stepName := fmt.Sprintf("Starting the agent '%s/%d'", i.jobName, i.id)
	err := stage.Perform(stepName, func() error {
		err := i.vm.Start()
		if err != nil {
			return bosherr.WrapError(err, "Starting the agent")
		}
		return nil
	})

	if err != nil {
		return err
	}

	err = i.waitUntilJobsAreRunning(update.UpdateWatchTime, stage)
	if err != nil {
		return err
	}

	if err := i.runScriptInStep(stage, "post-start", i.jobName, i.id); err != nil {
		return err
	}
	return nil
}

func (i *instance) shutdown(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	skipDrain bool,
	skipDiskUnmount bool,
	stage biui.Stage,
) error {
	if waitingForAgentErr := i.waitForAgent(pingDelay, pingTimeout, stage); waitingForAgentErr != nil {
		i.logger.Warn(i.logTag, "Gave up waiting for agent: %s", waitingForAgentErr.Error())
		return nil
	}

	if err := i.runScriptInStep(stage, "pre-stop", i.jobName, i.id); err != nil {
		return err
	}

	if !skipDrain {
		if err := i.drainJobs(stage); err != nil {
			return err
		}
	}

	if err := i.stopJobs(stage); err != nil {
		return err
	}

	if err := i.runScriptInStep(stage, "post-stop", i.jobName, i.id); err != nil {
		return err
	}

	if !skipDiskUnmount {
		if err := i.unmountDisks(stage); err != nil {
			return err
		}
	}

	return nil
}

func (i *instance) runScriptInStep(stage biui.Stage, script string, jobName string, jobId int) error {
	stepName := fmt.Sprintf("Running the %s scripts '%s/%d'", script, i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		err := i.vm.RunScript(script, map[string]interface{}{})
		if err != nil {
			msg := fmt.Sprintf("Running the %s script", script)
			return bosherr.WrapError(err, msg)
		}
		return nil
	})
}

func (i *instance) waitForAgent(pingDelay time.Duration, pingTimeout time.Duration, stage biui.Stage) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", i.vm.CID())
	return stage.Perform(stepName, func() error {
		if err := i.vm.WaitUntilReady(pingTimeout, pingDelay); err != nil {
			return bosherr.WrapError(err, "Agent unreachable")
		}
		return nil
	})
}

func (i *instance) waitUntilJobsAreRunning(updateWatchTime bideplmanifest.WatchTime, stage biui.Stage) error {
	start := time.Duration(updateWatchTime.Start) * time.Millisecond
	end := time.Duration(updateWatchTime.End) * time.Millisecond
	delayBetweenAttempts := 1 * time.Second
	maxAttempts := int((end - start) / delayBetweenAttempts)

	stepName := fmt.Sprintf("Waiting for instance '%s/%d' to be running", i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		time.Sleep(start)
		return i.vm.WaitToBeRunning(maxAttempts, delayBetweenAttempts)
	})
}

func (i *instance) drainJobs(stage biui.Stage) error {
	stepName := fmt.Sprintf("Draining jobs on instance '%s/%d'", i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		return i.vm.Drain()
	})
}

func (i *instance) stopJobs(stage biui.Stage) error {
	stepName := fmt.Sprintf("Stopping jobs on instance '%s/%d'", i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		return i.vm.Stop()
	})
}

func (i *instance) unmountDisks(stage biui.Stage) error {
	disks, err := i.vm.Disks()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting VM '%s' disks", i.vm.CID())
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Unmounting disk '%s'", disk.CID())
		err = stage.Perform(stepName, func() error {
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
