package disk

import (
	"reflect"
)

type Disk interface {
	CID() string
	NeedsMigration(newSize int, newCloudProperties map[string]interface{}) bool
}

type disk struct {
	cid             string
	size            int
	cloudProperties map[string]interface{}
}

func NewDisk(cid string, size int, cloudProperties map[string]interface{}) Disk {
	return &disk{
		cid:             cid,
		size:            size,
		cloudProperties: cloudProperties,
	}
}

func (d *disk) CID() string {
	return d.cid
}

func (d *disk) NeedsMigration(newSize int, newCloudProperties map[string]interface{}) bool {
	return d.size != newSize || !reflect.DeepEqual(d.cloudProperties, newCloudProperties)
}
