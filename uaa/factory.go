package uaa

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Factory struct {
	logTag string
	logger boshlog.Logger
}

func NewFactory(logger boshlog.Logger) Factory {
	return Factory{
		logTag: "uaa.Factory",
		logger: logger,
	}
}

func (f Factory) New(config Config) (UAA, error) {
	err := config.Validate()
	if err != nil {
		return UAAImpl{}, bosherr.WrapErrorf(
			err, "Validating UAA connection config")
	}

	client, err := f.httpClient(config)
	if err != nil {
		return UAAImpl{}, err
	}

	return UAAImpl{client: client}, nil
}

func (f Factory) httpClient(config Config) (Client, error) {
	certPool, err := config.CACertPool()
	if err != nil {
		return Client{}, err
	}

	if certPool == nil {
		f.logger.Debug(f.logTag, "Using default root CAs")
	} else {
		f.logger.Debug(f.logTag, "Using custom root CAs")
	}

	httpTransport := &http.Transport{
		TLSClientConfig:     &tls.Config{RootCAs: certPool},
		TLSHandshakeTimeout: 10 * time.Second,

		Dial:  (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 0}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	endpoint := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", config.Host, config.Port),
		User:   url.UserPassword(config.Client, config.ClientSecret),
	}

	rawClient := &http.Client{
		Transport: httpTransport,
	}

	httpClient := boshhttp.NewHTTPClient(rawClient, f.logger)

	return NewClient(endpoint.String(), httpClient, f.logger), nil
}
