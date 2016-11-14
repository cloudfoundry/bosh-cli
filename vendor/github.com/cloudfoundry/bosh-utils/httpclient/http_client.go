package httpclient

import (
	"net/http"
	"net"
	"strings"
	"crypto/x509"
	"crypto/tls"
	"fmt"
	"net/url"

	"errors"
	"regexp"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"net/url"
)

type HTTPClient interface {
	Post(endpoint string, payload []byte) (*http.Response, error)
	PostCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error)

	Put(endpoint string, payload []byte) (*http.Response, error)
	PutCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error)

	Get(endpoint string) (*http.Response, error)
	GetCustomized(endpoint string, f func(*http.Request)) (*http.Response, error)

	Delete(endpoint string) (*http.Response, error)

	VerifySSLCerts(endpoint string) SSLValidationError
}

// use interface to allow returning nil
type SSLValidationError interface {
  Cause() SSLValidationCause
  Error() string
  Endpoint() string
}
type sslError struct {
  cause SSLValidationCause
	err error
	endpoint string
}

func (s sslError) Cause() SSLValidationCause {
	return s.cause
}

func (s sslError) Error() string {
	return s.err.Error()
}

func (s sslError) Endpoint() string {
	return s.endpoint
}

type SSLValidationCause int
const (
  UnknownAuthorityError SSLValidationCause = iota
  CertNotValidForIPError SSLValidationCause = iota
  CertNotValidForHostnameError SSLValidationCause = iota
	UnknownValidationError SSLValidationCause = iota
)

type httpClient struct {
	client Client
	logger boshlog.Logger
	logTag string
	certPool *x509.CertPool
}

func NewHTTPClient(client Client, logger boshlog.Logger, certPool *x509.CertPool) HTTPClient {
	return httpClient{
		client: client,
		logger: logger,
		logTag: "httpClient",
		certPool: certPool,
	}
}

func (c httpClient) VerifySSLCerts(endpoint string) SSLValidationError {
	tlsConfig := &tls.Config{}
	if c.certPool != nil {
		tlsConfig.RootCAs = c.certPool
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return sslError{
			cause: UnknownValidationError,
			err: bosherr.WrapErrorf(err, fmt.Sprintf("Endpoint '%s' is not a valid URL")),
			endpoint: endpoint,
		}
	}
  host, port, err := net.SplitHostPort(endpointURL.Host)
	if err != nil {
	  host = endpointURL.Host
		port = "443"
	}
  tcpEndpoint := fmt.Sprintf("%s:%s", host, port)

	conn, err := tls.Dial("tcp", tcpEndpoint, tlsConfig)
	if _, ok := err.(x509.UnknownAuthorityError); ok {
		return sslError{
			cause: UnknownAuthorityError,
			err: err,
			endpoint: endpoint,
		}
	  // return fmt.Errorf("The SSL Certificate returned by '%s' is signed by an unknown Certificate Authority (CA). Please specify the CA Certificate with the `--ca-cert` flag.", tcpEndpoint)
	}
	if _, ok := err.(x509.HostnameError); ok {
	  if net.ParseIP(host) != nil {
			return sslError{
				cause: CertNotValidForIPError,
				err: err,
				endpoint: endpoint,
			}
		  // return fmt.Errorf("The SSL Certificate returned by '%s' is not valid for the IP address you specified: '%s'. Please specify the DNS Hostname instead of the IP with the `--environment` flag.", tcpEndpoint, err)
		} else {
			return sslError{
				cause: CertNotValidForHostnameError,
				err: err,
				endpoint: endpoint,
			}
		  // return fmt.Errorf("The SSL Certificate returned by '%s' did not match the hostname you specified: '%s'. Please specify a valid hostname or an IP address with the `--environment` flag.", tcpEndpoint, err)
		}
	}
	if err != nil {
		return sslError{
			cause: UnknownValidationError,
			err: bosherr.WrapErrorf(err, "Failed to verify certificate for endpoint '%s'", endpoint),
			endpoint: endpoint,
		}
	}
	conn.Close()

	return nil
}

func (c httpClient) Post(endpoint string, payload []byte) (*http.Response, error) {
	return c.PostCustomized(endpoint, payload, nil)
}

func (c httpClient) PostCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error) {
	postPayload := strings.NewReader(string(payload))

	c.logger.Debug(c.logTag, "Sending POST request to endpoint '%s' with body '%s'", scrubEndpointQuery(endpoint), payload)

	request, err := http.NewRequest("POST", endpoint, postPayload)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating POST request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(scrubErrorOutput(err), "Performing POST request")
	}

	return response, nil
}

func (c httpClient) Put(endpoint string, payload []byte) (*http.Response, error) {
	return c.PutCustomized(endpoint, payload, nil)
}

func (c httpClient) PutCustomized(endpoint string, payload []byte, f func(*http.Request)) (*http.Response, error) {
	putPayload := strings.NewReader(string(payload))

	c.logger.Debug(c.logTag, "Sending PUT request to endpoint '%s' with body '%s'", scrubEndpointQuery(endpoint), payload)

	request, err := http.NewRequest("PUT", endpoint, putPayload)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating PUT request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(scrubErrorOutput(err), "Performing PUT request")
	}

	return response, nil
}

func (c httpClient) Get(endpoint string) (*http.Response, error) {
	return c.GetCustomized(endpoint, nil)
}

func (c httpClient) GetCustomized(endpoint string, f func(*http.Request)) (*http.Response, error) {
	c.logger.Debug(c.logTag, "Sending GET request to endpoint '%s'", scrubEndpointQuery(endpoint))

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating GET request")
	}

	if f != nil {
		f(request)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(scrubErrorOutput(err), "Performing GET request")
	}

	return response, nil
}

func (c httpClient) Delete(endpoint string) (*http.Response, error) {
	c.logger.Debug(c.logTag, "Sending DELETE request with endpoint %s", endpoint)

	request, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating DELETE request")
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing DELETE request")
	}
	return response, nil
}

var scrubUserinfoRegex = regexp.MustCompile("(https?://.*:).*@")

func scrubEndpointQuery(endpoint string) string {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "error occurred parsing endpoing"
	}

	query := parsedURL.Query()
	query["refresh_token"] = []string{"<redacted>"}
	parsedURL.RawQuery = query.Encode()

	unescapedEndpoint, _ := url.QueryUnescape(parsedURL.String())
	return unescapedEndpoint
}

func scrubErrorOutput(err error) error {
	errorMsg := err.Error()
	errorMsg = scrubUserinfoRegex.ReplaceAllString(errorMsg, "$1<redacted>@")

	return errors.New(errorMsg)
}
