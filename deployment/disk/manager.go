package disk

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	bicloud "github.com/cloudfoundry/bosh-cli/v7/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/v7/config"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/v7/deployment/manifest"
	biui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

type Manager interface {
	// FindAllCurrent returns all disks currently associated with any VM in state.
	// Used by the deployment-level manager when collecting resources to delete.
	FindAllCurrent() ([]Disk, error)
	// FindCurrentForVM returns the disk currently associated with a specific VM.
	// Used by the disk deployer when deploying/attaching disks.
	FindCurrentForVM(vmCID string) ([]Disk, error)
	Create(bideplmanifest.DiskPool, string) (Disk, error)
	FindUnused() ([]Disk, error)
	DeleteUnused(biui.Stage) error
}

func NewManager(
	cloud bicloud.Cloud,
	diskRepo biconfig.DiskRepo,
	logger boshlog.Logger,
) Manager {
	return &manager{
		cloud:    cloud,
		diskRepo: diskRepo,
		logger:   logger,
		logTag:   "diskManager",
	}
}

type manager struct {
	cloud    bicloud.Cloud
	diskRepo biconfig.DiskRepo
	logger   boshlog.Logger
	logTag   string
}

func (m *manager) FindAllCurrent() ([]Disk, error) {
	allRecords, err := m.diskRepo.All()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting all disk records")
	}

	unusedRecords, err := m.diskRepo.FindUnused()
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting unused disk records")
	}

	unusedIDs := map[string]bool{}
	for _, r := range unusedRecords {
		unusedIDs[r.ID] = true
	}

	var disks []Disk
	for _, r := range allRecords {
		if !unusedIDs[r.ID] {
			disks = append(disks, NewDisk(r, m.cloud, m.diskRepo))
		}
	}
	return disks, nil
}

func (m *manager) FindCurrentForVM(vmCID string) ([]Disk, error) {
	disks := []Disk{}

	diskRecord, found, err := m.diskRepo.FindCurrentForVM(vmCID)
	if err != nil {
		return disks, bosherr.WrapError(err, "Reading disk record")
	}

	if found {
		disk := NewDisk(diskRecord, m.cloud, m.diskRepo)
		disks = append(disks, disk)
	}

	return disks, nil
}

func (m *manager) Create(diskPool bideplmanifest.DiskPool, vmCID string) (Disk, error) {
	diskCloudProperties := diskPool.CloudProperties

	m.logger.Debug(m.logTag, "Creating disk")
	cid, err := m.cloud.CreateDisk(diskPool.DiskSize, diskCloudProperties, vmCID)
	if err != nil {
		return nil,
			bosherr.WrapErrorf(err,
				"Creating disk with size %d, cloudProperties %#v, instanceID %s",
				diskPool.DiskSize, diskCloudProperties, vmCID,
			)
	}

	diskRecord, err := m.diskRepo.Save(cid, diskPool.DiskSize, diskCloudProperties)
	if err != nil {
		return nil, bosherr.WrapError(err, "Saving deployment disk record")
	}

	disk := NewDisk(diskRecord, m.cloud, m.diskRepo)

	return disk, nil
}

func (m *manager) FindUnused() ([]Disk, error) {
	disks := []Disk{}

	diskRecords, err := m.diskRepo.FindUnused()
	if err != nil {
		return disks, bosherr.WrapError(err, "Getting unused disk records")
	}

	for _, diskRecord := range diskRecords {
		disks = append(disks, NewDisk(diskRecord, m.cloud, m.diskRepo))
	}

	return disks, nil
}

func (m *manager) DeleteUnused(eventLoggerStage biui.Stage) error {
	disks, err := m.FindUnused()
	if err != nil {
		return bosherr.WrapError(err, "Finding unused disks")
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Deleting unused disk '%s'", disk.CID())
		err = eventLoggerStage.Perform(stepName, func() error {
			err := disk.Delete()
			var cloudErr bicloud.Error
			ok := errors.As(err, &cloudErr)
			if ok && cloudErr.Type() == bicloud.DiskNotFoundError {
				return biui.NewSkipStageError(cloudErr, "Disk Not Found")
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
