package director

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
		logTag: "director.Factory",
		logger: logger,
	}
}

func (f Factory) New(config Config, taskReporter TaskReporter, fileReporter FileReporter) (Director, error) {
	err := config.Validate()
	if err != nil {
		return DirectorImpl{}, bosherr.WrapErrorf(
			err, "Validating Director connection config")
	}

	client, err := f.httpClient(config, taskReporter, fileReporter)
	if err != nil {
		return DirectorImpl{}, err
	}

	return DirectorImpl{client: client}, nil
}

func (f Factory) httpClient(config Config, taskReporter TaskReporter, fileReporter FileReporter) (Client, error) {
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

	authAdjustment := NewAuthRequestAdjustment(
		config.TokenFunc, config.Username, config.Password)

	redirectFunc := func(req *http.Request, via []*http.Request) error {
		if len(via) > 10 {
			return bosherr.Error("Too many redirects")
		}

		// Since redirected requests are not retried,
		// forcefully adjust auth token as this is the last chance.
		err := authAdjustment.Adjust(req, true)
		if err != nil {
			return err
		}

		req.URL.Host = fmt.Sprintf("%s:%d", config.Host, config.Port)

		req.Header.Del("Referer")

		return nil
	}

	endpoint := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s:%d", config.Host, config.Port),
	}

	rawClient := &http.Client{
		Transport:     httpTransport,
		CheckRedirect: redirectFunc,
	}

	authedClient := NewAdjustableClient(rawClient, authAdjustment)

	httpClient := boshhttp.NewHTTPClient(authedClient, f.logger)

	return NewClient(endpoint.String(), httpClient, taskReporter, fileReporter, f.logger), nil
}
