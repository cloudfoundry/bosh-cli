package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
)

var DefaultClient = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

type HTTPClient interface {
	Post(string, []byte) (*http.Response, error)
	Get(string) (*http.Response, error)
}

type httpClient struct {
	client http.Client
	logger boshlog.Logger
	logTag string
}

func NewHTTPClient(logger boshlog.Logger) HTTPClient {
	return httpClient{
		client: DefaultClient,
		logger: logger,
		logTag: "httpClient",
	}
}

func (c httpClient) Post(endpoint string, payload []byte) (*http.Response, error) {
	postPayload := strings.NewReader(string(payload))
	c.logger.Debug(c.logTag, "Sending POST request with body %s, endpoint %s", payload, endpoint)

	request, err := http.NewRequest("POST", endpoint, postPayload)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating POST request")
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing POST request")
	}
	return response, nil
}

func (c httpClient) Get(endpoint string) (*http.Response, error) {
	response, err := http.Get(endpoint)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing GET request")
	}

	return response, nil
}
