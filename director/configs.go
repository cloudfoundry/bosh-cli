package director

import (
	"encoding/json"
	"fmt"
	"net/http"

	gourl "net/url"

	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	ID        string
	Name      string
	Type      string
	CreatedAt string `json:"created_at"`
	Team      string
	Content   string
	Current   bool `json:"current"`
}

type ConfigsFilter struct {
	Type string
	Name string
}

type UpdateConfigBody struct {
	Type             string `json:"type"`
	Name             string `json:"name"`
	Content          string `json:"content"`
	ExpectedLatestId string `json:"expected_latest_id,omitempty"`
}

type ConfigResponse struct {
	LatestId         string `json:"latest_id"`
	ExpectedLatestId string `json:"expected_latest_id"`
}

func (d DirectorImpl) LatestConfig(configType string, name string) (Config, error) {
	resps, err := d.client.latestConfig(configType, name)

	if err != nil {
		return Config{}, err
	}

	if len(resps) == 0 {
		return Config{}, bosherr.Error("No config")
	}

	return resps[0], nil
}

func (d DirectorImpl) LatestConfigByID(configID string) (Config, error) {
	return d.client.latestConfigByID(configID)
}

func (c Client) latestConfigByID(configID string) (Config, error) {
	var config Config

	path := fmt.Sprintf("/configs/%s", configID)

	respBody, response, err := c.clientRequest.RawGet(path, nil, nil)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return config, bosherr.WrapErrorf(err, "No config")
		}
		return config, bosherr.WrapErrorf(err, "Finding config")
	}

	err = json.Unmarshal(respBody, &config)

	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return config, nil
}

func (d DirectorImpl) ListConfigs(limit int, filter ConfigsFilter) ([]Config, error) {
	return d.client.listConfigs(limit, filter)
}

func (d DirectorImpl) UpdateConfig(configType string, name string, expectedLatestId string, content []byte) (Config, error) {
	body, err := json.Marshal(UpdateConfigBody{configType, name, string(content), expectedLatestId})
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Can't marshal request body")
	}
	return d.client.updateConfig(body)
}

func (d DirectorImpl) DeleteConfig(configType string, name string) (bool, error) {
	return d.client.deleteConfig(configType, name)
}

func (d DirectorImpl) DeleteConfigByID(configID string) (bool, error) {
	return d.client.deleteConfigByID(configID)
}

func (c Client) latestConfig(configType string, name string) ([]Config, error) {
	var resps []Config

	query := gourl.Values{}
	query.Add("type", configType)
	query.Add("name", name)
	query.Add("limit", "1")
	query.Add("latest", "true")
	path := fmt.Sprintf("/configs?%s", query.Encode())

	err := c.clientRequest.Get(path, &resps)
	if err != nil {
		return resps, bosherr.WrapError(err, "Finding config")
	}

	return resps, nil
}

func (c Client) listConfigs(limit int, filter ConfigsFilter) ([]Config, error) {
	var resps []Config

	query := gourl.Values{}
	if filter.Type != "" {
		query.Add("type", filter.Type)
	}
	if filter.Name != "" {
		query.Add("name", filter.Name)
	}
	query.Add("limit", fmt.Sprintf("%d", limit))
	query.Add("latest", strconv.FormatBool(limit == 1))
	path := fmt.Sprintf("/configs?%s", query.Encode())

	err := c.clientRequest.Get(path, &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Listing configs")
	}

	return resps, nil
}

func (c Client) updateConfig(content []byte) (Config, error) {
	var config Config

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	respBody, resp, err := c.clientRequest.RawPost("/configs", content, setHeaders)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusPreconditionFailed {
			var configResp ConfigResponse

			err = json.Unmarshal(respBody, &configResp)
			if err != nil {
				return config, bosherr.WrapErrorf(err, "Could not unmarshal response: '%s'", respBody)
			}

			return config, bosherr.Errorf("Config update rejected: The expected latest ID '%s' doesn't match the latest ID '%s'. This most likely means that a concurrent update of the config happened. Please try to upload again.", configResp.ExpectedLatestId, configResp.LatestId)
		}
		return config, bosherr.WrapErrorf(err, "Updating config")
	}

	err = json.Unmarshal(respBody, &config)
	if err != nil {
		return config, bosherr.WrapError(err, "Unmarshaling Director response")
	}

	return config, nil
}

func (c Client) deleteConfig(configType string, name string) (bool, error) {
	query := gourl.Values{}
	query.Add("type", configType)
	query.Add("name", name)
	path := fmt.Sprintf("/configs?%s", query.Encode())

	_, response, err := c.clientRequest.RawDelete(path)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, bosherr.WrapErrorf(err, "Deleting config")
	}

	return true, nil
}

func (c Client) deleteConfigByID(configID string) (bool, error) {
	path := fmt.Sprintf("/configs/%s", configID)

	_, response, err := c.clientRequest.RawDelete(path)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, bosherr.WrapErrorf(err, "Deleting config")
	}

	return true, nil
}

func (d DirectorImpl) DiffConfig(configType string, name string, manifest []byte) (ConfigDiff, error) {
	body, err := json.Marshal(UpdateConfigBody{configType, name, string(manifest), ""})
	if err != nil {
		return ConfigDiff{}, bosherr.WrapError(err, "Can't marshal request body")
	}
	resp, err := d.client.DiffConfig(body)
	if err != nil {
		return ConfigDiff{}, err
	}

	return NewConfigDiffWithFromId(resp.Diff, resp.From["id"]), nil
}

func (c Client) DiffConfig(manifest []byte) (ConfigDiffResponse, error) {
	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}
	return c.postConfigDiff("/configs/diff", manifest, setHeaders)
}
