package uaa

import (
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Client struct {
	clientRequest ClientRequest
}

func NewClient(endpoint string, httpClient boshhttp.HTTPClient, logger boshlog.Logger) Client {
	return Client{NewClientRequest(endpoint, httpClient, logger)}
}
