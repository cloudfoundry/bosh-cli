package director

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)
type VmResp struct {
	VmCID string `json:"vm_cid"`
}

func (d DeploymentImpl) DeleteVm(cid string) error {
	err := d.client.DeleteVm(cid)

	if err != nil {
		resps, listErr := d.client.Vms()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.VmCID == cid {
				return err
			}
		}
	}

	return nil
}

func (c Client) DeleteVm(cid string) error {
	if len(cid) == 0 {
		return bosherr.Error("Expected non-empty vm CID")
	}

	path := fmt.Sprintf("/vms/%s", cid)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Deleting vm '%s'", cid)
	}

	return nil
}

func (c Client) Vms() ([]VmResp, error) {
	var vms []VmResp

	err := c.clientRequest.Get("/vms", &vms)
	if err != nil {
		return vms, bosherr.WrapErrorf(
			err, "Listing vms")
	}

	return vms, nil
}