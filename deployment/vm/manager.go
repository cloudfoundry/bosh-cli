package vm

import (
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/clock"
	biagentclient "github.com/cloudfoundry/bosh-agent/v2/agentclient"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	bistemcell "github.com/cloudfoundry/bosh-cli/v7/stemcell"
)

type Manager interface {
	FindCurrent() (VM, bool, error)
	Create(stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest, diskCIDs []string) (VM, error)
}

type manager struct {
	vmRepo        biconfig.VMRepo
	stemcellRepo  biconfig.StemcellRepo
	diskDeployer  DiskDeployer
	agentClient   biagentclient.AgentClient
	cloud         bicloud.Cloud
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
	logger        boshlog.Logger
	logTag        string
	timeService   Clock
}

func NewManager(
	vmRepo biconfig.VMRepo,
	stemcellRepo biconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	agentClient biagentclient.AgentClient,
	cloud bicloud.Cloud,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
	timeService Clock,
) Manager {
	return &manager{
		cloud:         cloud,
		agentClient:   agentClient,
		vmRepo:        vmRepo,
		stemcellRepo:  stemcellRepo,
		diskDeployer:  diskDeployer,
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
		logTag:        "vmManager",
		timeService:   timeService,
	}
}

func (m *manager) FindCurrent() (VM, bool, error) {
	vmCID, found, err := m.vmRepo.FindCurrent()
	if err != nil {
		return nil, false, bosherr.WrapError(err, "Finding currently deployed vm")
	}

	if !found {
		return nil, false, nil
	}

	vm := NewVM(
		vmCID,
		m.vmRepo,
		m.stemcellRepo,
		m.diskDeployer,
		m.agentClient,
		m.cloud,
		clock.NewClock(),
		m.fs,
		m.logger,
	)

	return vm, true, err
}

func (m *manager) Create(stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest, diskCIDs []string) (VM, error) {
	jobName := deploymentManifest.JobName()
	networkInterfaces, err := deploymentManifest.NetworkInterfaces(jobName)
	m.logger.Debug(m.logTag, "Creating VM with network interfaces: %#v", networkInterfaces)
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting network spec")
	}

	resourcePool, err := deploymentManifest.ResourcePool(jobName)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting resource pool for job '%s'", jobName)
	}

	agentID, err := m.uuidGenerator.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating agent ID")
	}

	if len(deploymentManifest.Tags) > 0 {
		if _, ok := resourcePool.Env["bosh"]; ok {
			resourcePool.Env["bosh"].(biproperty.Map)["tags"] = deploymentManifest.Tags
		} else {
			resourcePool.Env["bosh"] = biproperty.Map{
				"tags": deploymentManifest.Tags,
			}
		}
	}
	cid, err := m.createAndRecordVM(agentID, stemcell, resourcePool, diskCIDs, networkInterfaces)
	if err != nil {
		return nil, err
	}

	metadata := bicloud.VMMetadata{
		"deployment":     deploymentManifest.Name,
		"job":            deploymentManifest.JobName(),
		"instance_group": deploymentManifest.JobName(),
		"index":          "0",
		"director":       "bosh-init",
		"name":           fmt.Sprintf("%s/%d", deploymentManifest.JobName(), 0),
		"created_at":     m.timeService.Now().Format(time.RFC3339),
	}

	for tagKey, tagValue := range deploymentManifest.Tags {
		metadata[tagKey] = tagValue
	}

	err = m.cloud.SetVMMetadata(cid, metadata)
	if err != nil {
		var cloudErr bicloud.Error
		ok := errors.As(err, &cloudErr)
		if ok && cloudErr.Type() == bicloud.NotImplementedError {
			// ignore it
		} else {
			return nil, bosherr.WrapErrorf(err, "Setting VM metadata to %s", metadata)
		}
	}

	vm := NewVMWithMetadata(
		cid,
		m.vmRepo,
		m.stemcellRepo,
		m.diskDeployer,
		m.agentClient,
		m.cloud,
		clock.NewClock(),
		m.fs,
		m.logger,
		metadata,
	)

	return vm, nil
}

func (m *manager) createAndRecordVM(agentID string, stemcell bistemcell.CloudStemcell, resourcePool bideplmanifest.ResourcePool, diskCIDs []string, networkInterfaces map[string]biproperty.Map) (string, error) {
	cid, err := m.cloud.CreateVM(agentID, stemcell.CID(), resourcePool.CloudProperties, diskCIDs, networkInterfaces, resourcePool.Env)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating vm with stemcell cid '%s'", stemcell.CID())
	}

	// Record vm info immediately so we don't leak it
	err = m.vmRepo.UpdateCurrent(cid)
	if err != nil {
		return "", bosherr.WrapError(err, "Updating current vm record")
	}

	return cid, nil
}
