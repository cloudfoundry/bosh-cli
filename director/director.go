package director

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DirectorImpl struct {
	client Client
}

func (d DirectorImpl) SetClient(c Client) DirectorImpl {
	d.client = c
	return d
}

type OrphanedVMResponse struct {
	AZName         string   `json:"az"`
	CID            string   `json:"cid"`
	DeploymentName string   `json:"deployment_name"`
	IPAddresses    []string `json:"ip_addresses"`
	InstanceName   string   `json:"instance_name"`
	OrphanedAt     string   `json:"orphaned_at"`
}

func (d DirectorImpl) WithContext(id string) Director {
	return DirectorImpl{client: d.client.WithContext(id)}
}

func (c Client) OrphanedVMs() ([]OrphanedVM, error) {
	var resps []OrphanedVMResponse

	err := c.clientRequest.Get("/orphaned_vms", &resps)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding orphaned VMs")
	}

	return transformOrphanedVMs(resps)
}

func transformOrphanedVMs(resps []OrphanedVMResponse) ([]OrphanedVM, error) {
	var orphanedVMs []OrphanedVM

	for _, r := range resps {
		orphanedAt, err := TimeParser{}.Parse(r.OrphanedAt)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Converting orphaned at '%s' to time", r.OrphanedAt)
		}

		orphanedVMs = append(orphanedVMs, OrphanedVM{
			CID:            r.CID,
			DeploymentName: r.DeploymentName,
			InstanceName:   r.InstanceName,
			AZName:         r.AZName,
			IPAddresses:    r.IPAddresses,
			OrphanedAt:     orphanedAt,
		})
	}
	return orphanedVMs, nil
}

func (d DirectorImpl) OrphanedVMs() ([]OrphanedVM, error) {
	return d.client.OrphanedVMs()
}

func (d DirectorImpl) EnableResurrection(enabled bool) error {
	return d.client.EnableResurrectionAll(enabled)
}

func (d DirectorImpl) DownloadResourceUnchecked(blobstoreID string, out io.Writer) error {
	return d.client.DownloadResourceUnchecked(blobstoreID, out)
}

func (c Client) EnableResurrectionAll(enabled bool) error {
	body := map[string]bool{"resurrection_paused": !enabled}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, _, err = c.clientRequest.RawPut("/resurrection", reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Changing VM resurrection state for all")
	}

	return nil
}

func (c Client) DownloadResourceUnchecked(blobstoreID string, out io.Writer) error {
	path := fmt.Sprintf("/resources/%s", blobstoreID)

	_, _, err := c.clientRequest.RawGet(path, out, nil)
	if err != nil {
		return bosherr.WrapErrorf(err, "Downloading resource '%s'", blobstoreID)
	}

	return nil
}

func (d DirectorImpl) CertificateExpiry() ([]CertificateExpiryInfo, error) {
	var resps []CertificateExpiryInfo
	responseBody, response, err := d.client.clientRequest.RawGet("/director/certificate_expiry", nil, nil)

	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			return nil, bosherr.WrapErrorf(err, "Certificate expiry information not supported")
		}
		return nil, bosherr.WrapErrorf(err, "Getting certificate expiry endpoint error")
	}

	err = json.Unmarshal(responseBody, &resps)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting certificate expiry endpoint error")
	}

	return resps, nil
}
