package instance

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployment/disk"
	bminstancestate "github.com/cloudfoundry/bosh-micro-cli/deployment/instance/state"
	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmsshtunnel "github.com/cloudfoundry/bosh-micro-cli/deployment/sshtunnel"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/deployment/vm"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Instance interface {
	JobName() string
	ID() int
	Disks() ([]bmdisk.Disk, error)
	WaitUntilReady(bminstallmanifest.Registry, bminstallmanifest.SSHTunnel, bmui.Stage) error
	UpdateDisks(bmdeplmanifest.Manifest, bmui.Stage) ([]bmdisk.Disk, error)
	UpdateJobs(bmdeplmanifest.Manifest, bmui.Stage) error
	Delete(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		stage bmui.Stage,
	) error
}

type instance struct {
	jobName          string
	id               int
	vm               bmvm.VM
	vmManager        bmvm.Manager
	sshTunnelFactory bmsshtunnel.Factory
	stateBuilder     bminstancestate.Builder
	logger           boshlog.Logger
	logTag           string
}

func NewInstance(
	jobName string,
	id int,
	vm bmvm.VM,
	vmManager bmvm.Manager,
	sshTunnelFactory bmsshtunnel.Factory,
	stateBuilder bminstancestate.Builder,
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

func (i *instance) Disks() ([]bmdisk.Disk, error) {
	disks, err := i.vm.Disks()
	if err != nil {
		return disks, bosherr.WrapError(err, "Listing instance disks")
	}
	return disks, nil
}

func (i *instance) WaitUntilReady(
	registryConfig bminstallmanifest.Registry,
	sshTunnelConfig bminstallmanifest.SSHTunnel,
	stage bmui.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s' to be ready", i.vm.CID())
	err := stage.Perform(stepName, func() error {
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

func (i *instance) UpdateDisks(deploymentManifest bmdeplmanifest.Manifest, stage bmui.Stage) ([]bmdisk.Disk, error) {
	diskPool, err := deploymentManifest.DiskPool(i.jobName)
	if err != nil {
		return []bmdisk.Disk{}, bosherr.WrapError(err, "Getting disk pool")
	}

	disks, err := i.vm.UpdateDisks(diskPool, stage)
	if err != nil {
		return disks, bosherr.WrapError(err, "Updating disks")
	}

	return disks, nil
}

func (i *instance) UpdateJobs(
	deploymentManifest bmdeplmanifest.Manifest,
	stage bmui.Stage,
) error {
	newState, err := i.stateBuilder.Build(i.jobName, i.id, deploymentManifest, stage)
	if err != nil {
		return bosherr.WrapErrorf(err, "Builing state for instance '%s/%d'", i.jobName, i.id)
	}

	stepName := fmt.Sprintf("Updating instance '%s/%d'", i.jobName, i.id)
	err = stage.Perform(stepName, func() error {
		err := i.vm.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping the agent")
		}

		err = i.vm.Apply(newState.ToApplySpec())
		if err != nil {
			return bosherr.WrapError(err, "Applying the agent state")
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

	return i.waitUntilJobsAreRunning(deploymentManifest.Update.UpdateWatchTime, stage)
}

func (i *instance) Delete(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	stage bmui.Stage,
) error {
	vmExists, err := i.vm.Exists()
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking existance of vm for instance '%s/%d'", i.jobName, i.id)
	}

	if vmExists {
		if err = i.shutdown(pingTimeout, pingDelay, stage); err != nil {
			return err
		}
	}

	// non-existent VMs still need to be 'deleted' to clean up related resources owned by the CPI
	stepName := fmt.Sprintf("Deleting VM '%s'", i.vm.CID())
	return stage.Perform(stepName, func() error {
		err := i.vm.Delete()
		cloudErr, ok := err.(bmcloud.Error)
		if ok && cloudErr.Type() == bmcloud.VMNotFoundError {
			return bmui.NewSkipStageError(cloudErr, "VM not found")
		}
		return err
	})
}

func (i *instance) shutdown(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	stage bmui.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", i.vm.CID())
	waitingForAgentErr := stage.Perform(stepName, func() error {
		if err := i.vm.WaitUntilReady(pingTimeout, pingDelay); err != nil {
			return bosherr.WrapError(err, "Agent unreachable")
		}
		return nil
	})
	if waitingForAgentErr != nil {
		i.logger.Warn(i.logTag, "Gave up waiting for agent: %s", waitingForAgentErr.Error())
		return nil
	}

	if err := i.stopJobs(stage); err != nil {
		return err
	}
	if err := i.unmountDisks(stage); err != nil {
		return err
	}
	return nil
}

func (i *instance) waitUntilJobsAreRunning(updateWatchTime bmdeplmanifest.WatchTime, stage bmui.Stage) error {
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

func (i *instance) stopJobs(stage bmui.Stage) error {
	stepName := fmt.Sprintf("Stopping jobs on instance '%s/%d'", i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		return i.vm.Stop()
	})
}

func (i *instance) unmountDisks(stage bmui.Stage) error {
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
