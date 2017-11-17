package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	proxy "github.com/cloudfoundry/socks5-proxy"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientFactory interface {
	CreateDefaultClient(*x509.CertPool) *http.Client
	CreateDefaultClientInsecureSkipVerify() *http.Client
}

type clientFactory struct {
	DefaultClient *http.Client
	defaultDialer DialFunc
}

func NewClientFactory() *clientFactory {
	// Calling this multiple times,
	// and calling the Dial method on separate instances of the dialer
	// has 1-2 seconds of performance overhead, so aim to have only
	// one clientFactory created per execution
	dialer := SOCKS5DialFuncFromEnvironment((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial, proxy.NewSocks5Proxy(proxy.NewHostKeyGetter()))

	return &clientFactory{
		DefaultClient: createDefaultClientInsecureSkipVerify(dialer),
		defaultDialer: dialer,
	}
}

func (c *clientFactory) CreateDefaultClient(certPool *x509.CertPool) *http.Client {
	insecureSkipVerify := false
	return newClient(insecureSkipVerify, certPool, c.defaultDialer)
}

func (c *clientFactory) CreateDefaultClientInsecureSkipVerify() *http.Client {
	insecureSkipVerify := true
	return newClient(insecureSkipVerify, nil, c.defaultDialer)
}

func createDefaultClient(certPool *x509.CertPool, dialer DialFunc) *http.Client {
	insecureSkipVerify := false
	return newClient(insecureSkipVerify, certPool, dialer)
}

func createDefaultClientInsecureSkipVerify(dialer DialFunc) *http.Client {
	insecureSkipVerify := true
	return newClient(insecureSkipVerify, nil, dialer)
}

func newClient(insecureSkipVerify bool, certPool *x509.CertPool, dialer DialFunc) *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
			TLSClientConfig: &tls.Config{
				RootCAs:            certPool,
				InsecureSkipVerify: insecureSkipVerify,
			},

			Proxy: http.ProxyFromEnvironment,
			Dial:  dialer,

			TLSHandshakeTimeout: 30 * time.Second,
			DisableKeepAlives:   true,
		},
	}

	return client
}
