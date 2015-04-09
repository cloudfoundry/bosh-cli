package disk

import (
	"reflect"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-init/cloud"
	bmproperty "github.com/cloudfoundry/bosh-init/common/property"
	bmconfig "github.com/cloudfoundry/bosh-init/config"
)

type Disk interface {
	CID() string
	NeedsMigration(newSize int, newCloudProperties bmproperty.Map) bool
	Delete() error
}

type disk struct {
	cid             string
	size            int
	cloudProperties bmproperty.Map

	cloud bmcloud.Cloud
	repo  bmconfig.DiskRepo
}

func NewDisk(
	diskRecord bmconfig.DiskRecord,
	cloud bmcloud.Cloud,
	repo bmconfig.DiskRepo,
) Disk {
	return &disk{
		cid:             diskRecord.CID,
		size:            diskRecord.Size,
		cloudProperties: diskRecord.CloudProperties,
		cloud:           cloud,
		repo:            repo,
	}
}

func (d *disk) CID() string {
	return d.cid
}

func (d *disk) NeedsMigration(newSize int, newCloudProperties bmproperty.Map) bool {
	return d.size != newSize || !reflect.DeepEqual(d.cloudProperties, newCloudProperties)
}

func (d *disk) Delete() error {
	deleteErr := d.cloud.DeleteDisk(d.cid)
	if deleteErr != nil {
		// allow DiskNotFoundError for idempotency
		cloudErr, ok := deleteErr.(bmcloud.Error)
		if !ok || cloudErr.Type() != bmcloud.DiskNotFoundError {
			return bosherr.WrapError(deleteErr, "Deleting disk in the cloud")
		}
	}

	diskRecord, found, err := d.repo.Find(d.cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding disk record (cid=%s)", d.cid)
	}

	if !found {
		return nil
	}

	err = d.repo.Delete(diskRecord)
	if err != nil {
		return bosherr.WrapError(err, "Deleting disk record")
	}

	// returns bmcloud.Error only if it is a DiskNotFoundError
	return deleteErr
}
