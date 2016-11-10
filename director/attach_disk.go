package director

import (
	"fmt"
	"net/http"
	"net/url"
)

func (d DirectorImpl) AttachDisk(deployment Deployment, instance InstanceSlug, diskCid string) error {
	return d.client.AttachDisk(deployment.Name(), instance, diskCid)
}

func (c Client) AttachDisk(deployment string, instance InstanceSlug, diskCid string) error {

	values := url.Values{}
	values.Add("deployment", deployment)
	values.Add("job", instance.Name())
	values.Add("instance_id", instance.IndexOrID())

	path := fmt.Sprintf("/disks/%s/attachments?%s", diskCid, values.Encode())
	_, err := c.taskClientRequest.PutResult(path, []byte{}, func(*http.Request) {})
	return err
}
