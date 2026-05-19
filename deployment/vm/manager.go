package vm

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"code.cloudfoundry.org/clock"
	bihttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http"
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

// ExistingVM pairs a running VM with the identity metadata stored in the repo.
type ExistingVM struct {
	VM         VM
	JobName    string
	InstanceID int
}

type Manager interface {
	FindAll() ([]ExistingVM, error)
	Create(jobName string, instanceID int, stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest, diskCIDs []string) (VM, error)
}

type manager struct {
	vmRepo             biconfig.VMRepo
	stemcellRepo       biconfig.StemcellRepo
	diskDeployer       DiskDeployer
	agentClientFactory bihttpagent.AgentClientFactory
	directorID         string
	mbusURL            string
	caCert             string
	cloud              bicloud.Cloud
	uuidGenerator      boshuuid.Generator
	fs                 boshsys.FileSystem
	logger             boshlog.Logger
	logTag             string
	timeService        Clock
}

func NewManager(
	vmRepo biconfig.VMRepo,
	stemcellRepo biconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	agentClientFactory bihttpagent.AgentClientFactory,
	directorID string,
	mbusURL string,
	caCert string,
	cloud bicloud.Cloud,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
	timeService Clock,
) Manager {
	return &manager{
		cloud:              cloud,
		agentClientFactory: agentClientFactory,
		directorID:         directorID,
		mbusURL:            mbusURL,
		caCert:             caCert,
		vmRepo:             vmRepo,
		stemcellRepo:       stemcellRepo,
		diskDeployer:       diskDeployer,
		uuidGenerator:      uuidGenerator,
		fs:                 fs,
		logger:             logger,
		logTag:             "vmManager",
		timeService:        timeService,
	}
}

func (m *manager) FindAll() ([]ExistingVM, error) {
	records, err := m.vmRepo.FindAll()
	if err != nil {
		return nil, bosherr.WrapError(err, "Finding currently deployed vms")
	}

	var existingVMs []ExistingVM
	for _, record := range records {
		mbusURL := record.MbusURL
		if mbusURL == "" {
			mbusURL = m.mbusURL
		}
		agentClient, err := m.agentClientFactory.NewAgentClient(m.directorID, mbusURL, m.caCert)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Creating agent client for VM '%s'", record.CID)
		}
		vm := NewVM(
			record.CID,
			mbusURL,
			m.vmRepo,
			m.stemcellRepo,
			m.diskDeployer,
			agentClient,
			m.cloud,
			clock.NewClock(),
			m.fs,
			m.logger,
		)
		existingVMs = append(existingVMs, ExistingVM{
			VM:         vm,
			JobName:    record.JobName,
			InstanceID: record.InstanceID,
		})
	}
	return existingVMs, nil
}

func (m *manager) Create(jobName string, instanceID int, stemcell bistemcell.CloudStemcell, deploymentManifest bideplmanifest.Manifest, diskCIDs []string) (VM, error) {
	networkInterfaces, err := deploymentManifest.NetworkInterfaces(jobName, instanceID)
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

	// Derive the per-instance mbus URL from the static IP for this instance.
	instanceMbusURL, err := m.agentMbusForInstance(jobName, instanceID, deploymentManifest)
	if err != nil {
		return nil, bosherr.WrapError(err, "Deriving per-instance mbus URL")
	}

	agentClient, err := m.agentClientFactory.NewAgentClient(m.directorID, instanceMbusURL, m.caCert)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating agent client for instance '%s/%d'", jobName, instanceID)
	}

	cid, err := m.createAndRecordVM(agentID, stemcell, resourcePool, diskCIDs, networkInterfaces, jobName, instanceID, instanceMbusURL)
	if err != nil {
		return nil, err
	}

	metadata := bicloud.VMMetadata{
		"deployment":     deploymentManifest.Name,
		"job":            jobName,
		"instance_group": jobName,
		"index":          strconv.Itoa(instanceID),
		"director":       "bosh-init",
		"name":           fmt.Sprintf("%s/%d", jobName, instanceID),
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
		instanceMbusURL,
		m.vmRepo,
		m.stemcellRepo,
		m.diskDeployer,
		agentClient,
		m.cloud,
		clock.NewClock(),
		m.fs,
		m.logger,
		metadata,
	)

	return vm, nil
}

func (m *manager) createAndRecordVM(
	agentID string,
	stemcell bistemcell.CloudStemcell,
	resourcePool bideplmanifest.ResourcePool,
	diskCIDs []string,
	networkInterfaces map[string]biproperty.Map,
	jobName string,
	instanceID int,
	mbusURL string,
) (string, error) {
	cid, err := m.cloud.CreateVM(agentID, stemcell.CID(), resourcePool.CloudProperties, diskCIDs, networkInterfaces, resourcePool.Env)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating vm with stemcell cid '%s'", stemcell.CID())
	}

	// Record vm info immediately so we don't leak it.
	_, err = m.vmRepo.Save(jobName, instanceID, cid, mbusURL)
	if err != nil {
		return "", bosherr.WrapError(err, "Saving vm record")
	}

	return cid, nil
}

// agentMbusForInstance derives the per-instance agent mbus URL.
// It substitutes the instance's static IP (at index instanceID) into the base
// mbus URL from cloud_provider. When no static IPs are configured (dynamic
// networking), the base mbus URL is returned unchanged for backward compatibility.
func (m *manager) agentMbusForInstance(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest) (string, error) {
	staticIP := m.staticIPForInstance(jobName, instanceID, deploymentManifest)
	if staticIP == "" {
		return m.mbusURL, nil
	}
	return replaceHost(m.mbusURL, staticIP)
}

func (m *manager) staticIPForInstance(jobName string, instanceID int, deploymentManifest bideplmanifest.Manifest) string {
	job, found := deploymentManifest.FindJobByName(jobName)
	if !found {
		return ""
	}
	for _, jn := range job.Networks {
		if len(jn.StaticIPs) > instanceID {
			return jn.StaticIPs[instanceID]
		}
	}
	return ""
}

// replaceHost substitutes newHost into the host portion of rawURL, keeping the
// original port and credentials.
func replaceHost(rawURL, newHost string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Parsing mbus URL '%s'", rawURL)
	}
	_, port, splitErr := net.SplitHostPort(u.Host)
	if splitErr != nil {
		// No port in URL — just replace the host as-is.
		u.Host = newHost
	} else {
		u.Host = net.JoinHostPort(newHost, port)
	}
	return u.String(), nil
}
