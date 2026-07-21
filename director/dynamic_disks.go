package director

import (
	"fmt"
	"net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DynamicDiskImpl struct {
	name             string
	diskCID          string
	deploymentName   string
	instanceName     string
	availabilityZone string
	size             uint64
	diskPoolName     string
	cpi              string
}

func (d DynamicDiskImpl) Name() string             { return d.name }
func (d DynamicDiskImpl) DiskCID() string          { return d.diskCID }
func (d DynamicDiskImpl) DeploymentName() string   { return d.deploymentName }
func (d DynamicDiskImpl) InstanceName() string     { return d.instanceName }
func (d DynamicDiskImpl) AvailabilityZone() string { return d.availabilityZone }
func (d DynamicDiskImpl) Size() uint64             { return d.size }
func (d DynamicDiskImpl) DiskPoolName() string     { return d.diskPoolName }
func (d DynamicDiskImpl) CPI() string              { return d.cpi }

type DynamicDiskResp struct {
	Name             string `json:"name"`
	DiskCID          string `json:"disk_cid"`
	Deployment       string `json:"deployment"`
	Instance         string `json:"instance"`
	AvailabilityZone string `json:"availability_zone"`
	Size             uint64 `json:"size"`
	DiskPoolName     string `json:"disk_pool_name"`
	CPI              string `json:"cpi"`
}

// --- DirectorImpl delegation ---

func (d DirectorImpl) DeleteDynamicDisk(diskName string) error {
	return d.client.DeleteDynamicDisk(diskName)
}

func (d DirectorImpl) DynamicDisks() ([]DynamicDisk, error) {
	return d.client.DynamicDisks()
}

// --- HTTP Client methods ---

func (c Client) DeleteDynamicDisk(diskName string) error {
	if len(diskName) == 0 {
		return bosherr.Error("Expected non-empty dynamic disk name")
	}

	path := fmt.Sprintf("/dynamic_disks/%s", url.PathEscape(diskName))
	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting dynamic disk '%s'", diskName)
	}
	return nil
}

func (c Client) DynamicDisks() ([]DynamicDisk, error) {
	var resps []DynamicDiskResp
	if err := c.clientRequest.Get("/dynamic_disks", &resps); err != nil {
		return nil, bosherr.WrapError(err, "Listing dynamic disks")
	}

	var disks []DynamicDisk
	for _, r := range resps {
		disks = append(disks, DynamicDiskImpl{
			name:             r.Name,
			diskCID:          r.DiskCID,
			deploymentName:   r.Deployment,
			instanceName:     r.Instance,
			availabilityZone: r.AvailabilityZone,
			size:             r.Size,
			diskPoolName:     r.DiskPoolName,
			cpi:              r.CPI,
		})
	}
	return disks, nil
}
