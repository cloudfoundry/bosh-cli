package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"code.cloudfoundry.org/tlsconfig"

	proxy "github.com/cloudfoundry/socks5-proxy"
)

var (
	DefaultClient            = CreateDefaultClientInsecureSkipVerify()
	defaultDialerContextFunc = SOCKS5DialContextFuncFromEnvironment((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}), proxy.NewSocks5Proxy(proxy.NewHostKey(), log.New(ioutil.Discard, "", log.LstdFlags), 1*time.Minute))
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

func CreateDefaultClient(certPool *x509.CertPool) *http.Client {
	insecureSkipVerify := false
	external := false
	disableKeepAlives := true
	return factory{}.New(insecureSkipVerify, external, disableKeepAlives, certPool)
}

func CreateExternalDefaultClient(certPool *x509.CertPool) *http.Client {
	insecureSkipVerify := false
	external := true
	disableKeepAlives := true
	return factory{}.New(insecureSkipVerify, external, disableKeepAlives, certPool)
}

func CreateKeepAliveDefaultClient(certPool *x509.CertPool) *http.Client {
	insecureSkipVerify := false
	external := true
	disableKeepAlives := false
	return factory{}.New(insecureSkipVerify, external, disableKeepAlives, certPool)
}

func CreateDefaultClientInsecureSkipVerify() *http.Client {
	insecureSkipVerify := true
	external := false
	disableKeepAlives := true
	return factory{}.New(insecureSkipVerify, external, disableKeepAlives, nil)
}

type factory struct{}

func (f factory) New(insecureSkipVerify, externalClient bool, disableKeepAlives bool, certPool *x509.CertPool) *http.Client {
	serviceDefaults := tlsconfig.WithInternalServiceDefaults()
	if externalClient {
		serviceDefaults = tlsconfig.WithExternalServiceDefaults()
	}

	tlsConfig, err := tlsconfig.Build(
		serviceDefaults,
		WithInsecureSkipVerify(insecureSkipVerify),
		WithClientSessionCache(0),
	).Client(tlsconfig.WithAuthority(certPool))
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsConfig,
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         defaultDialerContextFunc,
			TLSHandshakeTimeout: 30 * time.Second,
			DisableKeepAlives:   disableKeepAlives,
		},
	}

	return client
}

func WithInsecureSkipVerify(insecureSkipVerify bool) tlsconfig.TLSOption {
	return func(config *tls.Config) error {
		config.InsecureSkipVerify = insecureSkipVerify
		return nil
	}
}

func WithClientSessionCache(capacity int) tlsconfig.TLSOption {
	return func(config *tls.Config) error {
		config.ClientSessionCache = tls.NewLRUClientSessionCache(capacity)
		return nil
	}
}
