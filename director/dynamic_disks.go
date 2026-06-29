package director

import (
	"encoding/json"
	"fmt"
	"net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DynamicDiskImpl struct {
	client Client

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

type ProvideDynamicDiskResult struct {
	DiskCID string `json:"disk_cid"`
}

// --- DirectorImpl delegation ---

func (d DirectorImpl) ProvideDynamicDisk(instanceID, diskName, diskPool string, sizeInMB int, metadata map[string]interface{}) (string, error) {
	return d.client.ProvideDynamicDisk(instanceID, diskName, diskPool, sizeInMB, metadata)
}

func (d DirectorImpl) DetachDynamicDisk(diskName string) error {
	return d.client.DetachDynamicDisk(diskName)
}

func (d DirectorImpl) DeleteDynamicDisk(diskName string) error {
	return d.client.DeleteDynamicDisk(diskName)
}

func (d DirectorImpl) DynamicDisks() ([]DynamicDisk, error) {
	return d.client.DynamicDisks()
}

func (d DirectorImpl) CreateDynamicDisk(diskName, diskPool string, sizeInMB int, metadata map[string]interface{}) (string, error) {
	return d.client.CreateDynamicDisk(diskName, diskPool, sizeInMB, metadata)
}

func (d DirectorImpl) AttachDynamicDisk(diskName, instanceID string) error {
	return d.client.AttachDynamicDisk(diskName, instanceID)
}

// --- HTTP Client methods ---

func (c Client) ProvideDynamicDisk(instanceID, diskName, diskPool string, sizeInMB int, metadata map[string]interface{}) (string, error) {
	reqBody := map[string]interface{}{
		"instance_id":    instanceID,
		"disk_name":      diskName,
		"disk_pool_name": diskPool,
		"disk_size":      sizeInMB,
	}
	if metadata != nil {
		reqBody["metadata"] = metadata
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", bosherr.WrapError(err, "Marshaling provide dynamic disk request")
	}

	resultBytes, err := c.taskClientRequest.PostResult("/dynamic_disks/provide", bodyBytes, nil)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Providing dynamic disk '%s'", diskName)
	}

	var result ProvideDynamicDiskResult
	if len(resultBytes) > 0 {
		if parseErr := json.Unmarshal(resultBytes, &result); parseErr != nil {
			return "", bosherr.WrapErrorf(parseErr, "Unmarshaling provide disk result")
		}
	}
	return result.DiskCID, nil
}

func (c Client) DetachDynamicDisk(diskName string) error {
	path := fmt.Sprintf("/dynamic_disks/%s/detach", url.PathEscape(diskName))
	_, err := c.taskClientRequest.PostResult(path, nil, nil)
	if err != nil {
		return bosherr.WrapErrorf(err, "Detaching dynamic disk '%s'", diskName)
	}
	return nil
}

func (c Client) DeleteDynamicDisk(diskName string) error {
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
			client:           c,
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

func (c Client) CreateDynamicDisk(diskName, diskPool string, sizeInMB int, metadata map[string]interface{}) (string, error) {
	reqBody := map[string]interface{}{
		"disk_name":      diskName,
		"disk_pool_name": diskPool,
		"disk_size":      sizeInMB,
	}
	if metadata != nil {
		reqBody["metadata"] = metadata
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", bosherr.WrapError(err, "Marshaling create dynamic disk request")
	}

	resultBytes, err := c.taskClientRequest.PostResult("/dynamic_disks", bodyBytes, nil)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating dynamic disk '%s'", diskName)
	}

	var result ProvideDynamicDiskResult
	if len(resultBytes) > 0 {
		if parseErr := json.Unmarshal(resultBytes, &result); parseErr != nil {
			return "", bosherr.WrapErrorf(parseErr, "Unmarshaling create disk result")
		}
	}
	return result.DiskCID, nil
}

func (c Client) AttachDynamicDisk(diskName, instanceID string) error {
	reqBody := map[string]interface{}{
		"instance_id": instanceID,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling attach dynamic disk request")
	}

	path := fmt.Sprintf("/dynamic_disks/%s/attach", url.PathEscape(diskName))
	_, err = c.taskClientRequest.PostResult(path, bodyBytes, nil)
	if err != nil {
		return bosherr.WrapErrorf(err, "Attaching dynamic disk '%s' to instance '%s'", diskName, instanceID)
	}
	return nil
}
