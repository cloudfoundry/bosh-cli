package director

import (
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TaskConfig struct {
	Properties string
}

func (d DirectorImpl) LatestTaskConfig() (TaskConfig, error) {
	resps, err := d.client.TaskConfigs()
	if err != nil {
		return TaskConfig{}, err
	}

	if len(resps) == 0 {
		return TaskConfig{}, bosherr.Error("No Task config")
	}

	return resps[0], nil
}

func (d DirectorImpl) UpdateTaskConfig(manifest []byte) error {
	return d.client.UpdateTaskConfig(manifest)
}

func (c Client) TaskConfigs() ([]TaskConfig, error) {
	var resps []TaskConfig

	err := c.clientRequest.Get("/task_configs?limit=1", &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Finding Task configs")
	}

	return resps, nil
}

func (c Client) UpdateTaskConfig(manifest []byte) error {
	path := "/task_configs"

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	_, _, err := c.clientRequest.RawPost(path, manifest, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating Task config")
	}

	return nil
}
