package vm

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmretrystrategy "github.com/cloudfoundry/bosh-micro-cli/deployer/retrystrategy"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

type VM interface {
	CID() string
	WaitToBeReady(timeout time.Duration, delay time.Duration) error
	Apply(bmstemcell.ApplySpec, bmdepl.Deployment) error
	Start() error
	WaitToBeRunning(maxAttempts int, delay time.Duration) error
	AttachDisk(bmdisk.Disk) error
	DetachDisk(bmdisk.Disk) error
	Stop() error
	Disks() ([]bmdisk.Disk, error)
	UnmountDisk(bmdisk.Disk) error
	MigrateDisk() error
	Delete() error
}

type vm struct {
	cid                    string
	vmRepo                 bmconfig.VMRepo
	agentClient            bmagentclient.AgentClient
	cloud                  bmcloud.Cloud
	templatesSpecGenerator bmas.TemplatesSpecGenerator
	applySpecFactory       bmas.Factory
	mbusURL                string
	fs                     boshsys.FileSystem
	logger                 boshlog.Logger
	logTag                 string
}

func NewVM(
	cid string,
	vmRepo bmconfig.VMRepo,
	agentClient bmagentclient.AgentClient,
	cloud bmcloud.Cloud,
	templatesSpecGenerator bmas.TemplatesSpecGenerator,
	applySpecFactory bmas.Factory,
	mbusURL string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) VM {
	return &vm{
		cid:         cid,
		vmRepo:      vmRepo,
		agentClient: agentClient,
		cloud:       cloud,
		templatesSpecGenerator: templatesSpecGenerator,
		applySpecFactory:       applySpecFactory,
		mbusURL:                mbusURL,
		fs:                     fs,
		logger:                 logger,
		logTag:                 "vm",
	}
}

func (vm *vm) CID() string {
	return vm.cid
}

func (vm *vm) WaitToBeReady(timeout time.Duration, delay time.Duration) error {
	agentPingRetryable := bmagentclient.NewPingRetryable(vm.agentClient)
	timeService := boshtime.NewConcreteService()
	agentPingRetryStrategy := bmretrystrategy.NewTimeoutRetryStrategy(timeout, delay, agentPingRetryable, timeService, vm.logger)
	return agentPingRetryStrategy.Try()
}

func (vm *vm) Apply(stemcellApplySpec bmstemcell.ApplySpec, deployment bmdepl.Deployment) error {
	vm.logger.Debug(vm.logTag, "Stopping agent")

	err := vm.agentClient.Stop()
	if err != nil {
		return bosherr.WrapError(err, "Stopping agent")
	}

	vm.logger.Debug(vm.logTag, "Rendering job templates")
	renderedJobDir, err := vm.fs.TempDir("instance-updater-render-job")
	if err != nil {
		return bosherr.WrapError(err, "Creating rendered job directory")
	}
	defer vm.fs.RemoveAll(renderedJobDir)

	deploymentJob := deployment.Jobs[0]
	jobProperties, err := deploymentJob.Properties()
	if err != nil {
		return bosherr.WrapError(err, "Stringifying job properties")
	}

	networksSpec, err := deployment.NetworksSpec(deploymentJob.Name)
	if err != nil {
		return bosherr.WrapError(err, "Stringifying job properties")
	}

	templatesSpec, err := vm.templatesSpecGenerator.Create(
		deploymentJob,
		stemcellApplySpec.Job,
		deployment.Name,
		jobProperties,
		vm.mbusURL,
	)
	if err != nil {
		return bosherr.WrapError(err, "Generating templates spec")
	}

	vm.logger.Debug(vm.logTag, "Creating apply spec")
	agentApplySpec := vm.applySpecFactory.Create(
		stemcellApplySpec,
		deployment.Name,
		deploymentJob.Name,
		networksSpec,
		templatesSpec.BlobID,
		templatesSpec.ArchiveSha1,
		templatesSpec.ConfigurationHash,
	)

	vm.logger.Debug(vm.logTag, "Sending apply message to the agent with '%#v'", agentApplySpec)
	err = vm.agentClient.Apply(agentApplySpec)
	if err != nil {
		return bosherr.WrapError(err, "Sending apply spec to agent")
	}

	return nil
}

func (vm *vm) Start() error {
	return vm.agentClient.Start()
}

func (vm *vm) WaitToBeRunning(maxAttempts int, delay time.Duration) error {
	agentGetStateRetryable := bmagentclient.NewGetStateRetryable(vm.agentClient)
	agentGetStateRetryStrategy := bmretrystrategy.NewAttemptRetryStrategy(maxAttempts, delay, agentGetStateRetryable, vm.logger)
	return agentGetStateRetryStrategy.Try()
}

func (vm *vm) AttachDisk(disk bmdisk.Disk) error {
	err := vm.cloud.AttachDisk(vm.cid, disk.CID())
	if err != nil {
		return bosherr.WrapError(err, "Attaching disk in the cloud")
	}

	err = vm.agentClient.MountDisk(disk.CID())
	if err != nil {
		return bosherr.WrapError(err, "Mounting disk")
	}

	return nil
}

func (vm *vm) DetachDisk(disk bmdisk.Disk) error {
	err := vm.cloud.DetachDisk(vm.cid, disk.CID())
	if err != nil {
		return bosherr.WrapError(err, "Detaching disk in the cloud")
	}

	return nil
}

func (vm *vm) Stop() error {
	return vm.agentClient.Stop()
}

func (vm *vm) Disks() ([]bmdisk.Disk, error) {
	result := []bmdisk.Disk{}

	disks, err := vm.agentClient.ListDisk()
	if err != nil {
		return result, bosherr.WrapError(err, "Listing vm disks")
	}

	for _, diskCID := range disks {
		result = append(result, bmdisk.NewDisk(diskCID, 0, map[string]interface{}{}))
	}

	return result, nil
}

func (vm *vm) UnmountDisk(disk bmdisk.Disk) error {
	return vm.agentClient.UnmountDisk(disk.CID())
}

func (vm *vm) MigrateDisk() error {
	return vm.agentClient.MigrateDisk()
}

func (vm *vm) Delete() error {
	err := vm.cloud.DeleteVM(vm.cid)
	if err != nil {
		return bosherr.WrapError(err, "Deleting vm in the cloud")
	}

	err = vm.vmRepo.ClearCurrent()
	if err != nil {
		return bosherr.WrapError(err, "Deleting vm from vm repo")
	}

	return nil
}
