package disk

import (
	"reflect"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
)

type Disk interface {
	CID() string
	NeedsMigration(newSize int, newCloudProperties map[string]interface{}) bool
	Delete() error
}

type disk struct {
	cid             string
	size            int
	cloudProperties map[string]interface{}

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

func (d *disk) NeedsMigration(newSize int, newCloudProperties map[string]interface{}) bool {
	return d.size != newSize || !reflect.DeepEqual(d.cloudProperties, newCloudProperties)
}

func (d *disk) Delete() error {
	err := d.cloud.DeleteDisk(d.cid)
	if err != nil {
		return bosherr.WrapError(err, "Deleting disk from cloud")
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

	return nil
}
