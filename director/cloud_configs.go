package director

import (
	"fmt"
	"net/http"
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CloudConfig struct {
	Properties string
}

func (d DirectorImpl) LatestCloudConfig(name string) (CloudConfig, error) {
	resps, err := d.client.CloudConfigs(name)
	if err != nil {
		return CloudConfig{}, err
	}

	if len(resps) == 0 {
		return CloudConfig{}, bosherr.Error("No cloud config")
	}

	return resps[0], nil
}

func (d DirectorImpl) UpdateCloudConfig(name string, manifest []byte) error {
	return d.client.UpdateCloudConfig(name, manifest)
}

func (c Client) CloudConfigs(name string) ([]CloudConfig, error) {
	var resps []CloudConfig

	query := gourl.Values{}
	query.Add("name", name)
	query.Add("limit", "1")

	path := fmt.Sprintf("/cloud_configs?%s", query.Encode())

	err := c.clientRequest.Get(path, &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Finding cloud configs")
	}

	return resps, nil
}

func (c Client) UpdateCloudConfig(name string, manifest []byte) error {
	query := gourl.Values{}
	query.Add("name", name)

	path := fmt.Sprintf("/cloud_configs?%s", query.Encode())

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	_, _, err := c.clientRequest.RawPost(path, manifest, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating cloud config")
	}

	return nil
}

func (d DirectorImpl) DiffCloudConfig(name string, manifest []byte) (ConfigDiff, error) {
	resp, err := d.client.DiffCloudConfig(name, manifest)
	if err != nil {
		return ConfigDiff{}, err
	}

	return NewConfigDiff(resp.Diff), nil
}

func (c Client) DiffCloudConfig(name string, manifest []byte) (ConfigDiffResponse, error) {
	query := gourl.Values{}
	query.Add("name", name)

	path := fmt.Sprintf("/cloud_configs/diff?%s", query.Encode())

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	return c.postConfigDiff(path, manifest, setHeaders)
}
