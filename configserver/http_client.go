package configserver

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"path/filepath"
	"time"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshcry "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshhttp "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type HTTPClientOpts struct {
	URL string

	UAAURL          string
	UAAClient       string
	UAAClientSecret string

	TLSCA          []byte
	TLSCertificate []byte
	TLSPrivateKey  []byte

	Namespace string
}

type HTTPClient struct {
	endpoint   string
	namespace  string
	httpClient *boshhttp.HTTPClient
	logger     boshlog.Logger
}

var _ Client = HTTPClient{}

func NewHTTPClient(opts HTTPClientOpts, logger boshlog.Logger) (HTTPClient, error) {
	if len(opts.URL) == 0 {
		return HTTPClient{}, bosherr.Errorf("Expected config server URL to be non-empty")
	}

	if len(opts.Namespace) == 0 {
		return HTTPClient{}, bosherr.Errorf("Expected config server namespace to be non-empty")
	}

	var caCertPool *x509.CertPool

	if len(opts.TLSCA) > 0 {
		var err error
		caCertPool, err = boshcry.CertPoolFromPEM(opts.TLSCA)
		if err != nil {
			return HTTPClient{}, bosherr.WrapErrorf(err, "Building config server CA certificate")
		}
	}

	var client boshhttp.Client

	switch {
	case len(opts.UAAURL) > 0:
		uaaConfig, err := boshuaa.NewConfigFromURL(opts.UAAURL)
		if err != nil {
			return HTTPClient{}, err
		}

		uaaConfig.CACert = string(opts.TLSCA)
		uaaConfig.Client = opts.UAAClient
		uaaConfig.ClientSecret = opts.UAAClientSecret

		uaa, err := boshuaa.NewFactory(logger).New(uaaConfig)
		if err != nil {
			return HTTPClient{}, bosherr.WrapErrorf(err, "Building config server UAA")
		}

		rawClient := boshhttp.CreateDefaultClient(caCertPool)
		retryClient := boshhttp.NewNetworkSafeRetryClient(rawClient, 5, 500*time.Millisecond, logger)
		authAdjustment := boshdir.NewAuthRequestAdjustment(boshuaa.NewClientTokenSession(uaa).TokenFunc, "", "")
		client = boshdir.NewAdjustableClient(retryClient, authAdjustment)

	case len(opts.TLSCertificate) > 0:
		cert, err := tls.X509KeyPair(opts.TLSCertificate, opts.TLSPrivateKey)
		if err != nil {
			return HTTPClient{}, bosherr.WrapErrorf(err, "Building config server client certificate")
		}
		client = boshhttp.NewMutualTLSClient(cert, caCertPool, "")
		client = boshhttp.NewNetworkSafeRetryClient(client, 4, 500*time.Millisecond, logger)

	default:
		return HTTPClient{}, bosherr.Errorf("Expected non-empty config server authentication details")
	}

	httpClient := boshhttp.NewHTTPClient(client, logger)

	return HTTPClient{opts.URL, opts.Namespace, httpClient, logger}, nil
}

type configValueData struct {
	Data []configValue
}

type configValue struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type configValueWrite struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Mode  string      `json:"mode"`
	Value interface{} `json:"value"`
}

type configValueTemplate struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Parameters interface{} `json:"parameters"`
}

func (c HTTPClient) Read(name string) (interface{}, error) {
	name = c.namespacedName(name)

	var data configValueData

	query := gourl.Values{}
	query.Add("name", name)

	_, err := c.get("/api/v1/data?"+query.Encode(), &data)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading config server value '%s'", name)
	}

	if len(data.Data) == 0 {
		return nil, bosherr.Errorf("Expected to find at least one config server value for '%s'", name)
	}

	return c.putInterfaceKeys(data.Data[0].Value), nil
}

func (c HTTPClient) Exists(name string) (bool, error) {
	name = c.namespacedName(name)

	var data configValueData

	query := gourl.Values{}
	query.Add("name", name)

	resp, err := c.get("/api/v1/data?"+query.Encode(), &data)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, bosherr.WrapErrorf(err, "Reading config server value '%s'", name)
	}

	return len(data.Data) > 0, nil
}

func (c HTTPClient) Write(name string, value interface{}) error {
	name = c.namespacedName(name)

	// unfortunately there is no better way to encode []byte
	// could potentially do it via custom type in config server
	// currently only support top level []byte values
	if bytes, ok := value.([]byte); ok {
		value = string(bytes)
	}

	tpl := configValueWrite{
		Name:  name,
		Type:  "value",
		Mode:  "overwrite",
		Value: c.stripInterfaceKeys(value),
	}

	tplBytes, err := json.Marshal(tpl)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling config server value")
	}

	var val configValue

	err = c.put("/api/v1/data", tplBytes, &val)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing config server value '%s'", name)
	}

	return nil
}

func (c HTTPClient) Delete(name string) error {
	name = c.namespacedName(name)

	query := gourl.Values{}
	query.Add("name", name)

	err := c.delete("/api/v1/data?" + query.Encode())
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting config server value '%s'", name)
	}

	return nil
}

func (c HTTPClient) Generate(name, type_ string, params interface{}) (interface{}, error) {
	name = c.namespacedName(name)

	tpl := configValueTemplate{
		Name:       name,
		Type:       type_,
		Parameters: c.namespaceCAIfNecessary(c.stripInterfaceKeys(params)),
	}

	tplBytes, err := json.Marshal(tpl)
	if err != nil {
		return nil, bosherr.WrapError(err, "Marshaling config server value")
	}

	var val configValue

	err = c.post("/api/v1/data", tplBytes, &val)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Generating config server value '%s'", name)
	}

	return c.putInterfaceKeys(val.Value), nil
}

func (r HTTPClient) get(path string, response interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
	}

	resp, err := r.httpClient.GetCustomized(url, setHeaders)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Performing request GET '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return resp, err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return resp, bosherr.WrapError(err, "Unmarshaling config server response")
	}

	return resp, nil
}

func (r HTTPClient) post(path string, payload []byte, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := r.httpClient.PostCustomized(url, payload, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request POST '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling config server response")
	}

	return nil
}

func (r HTTPClient) put(path string, payload []byte, response interface{}) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := r.httpClient.PutCustomized(url, payload, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request PUT '%s'", url)
	}

	respBody, err := r.readResponse(resp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return bosherr.WrapError(err, "Unmarshaling config server response")
	}

	return nil
}

func (r HTTPClient) delete(path string) error {
	url := fmt.Sprintf("%s%s", r.endpoint, path)

	setHeaders := func(req *http.Request) {
		req.Header.Add("Accept", "application/json")
	}

	resp, err := r.httpClient.DeleteCustomized(url, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Performing request DELETE '%s'", url)
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusNoContent {
		msg := "Config server responded with non-successful status code '%d'"
		return bosherr.Errorf(msg, resp.StatusCode)
	}

	return nil
}

func (r HTTPClient) readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading config server response")
	}

	if resp.StatusCode != http.StatusOK {
		msg := "Config server responded with non-successful status code '%d' response '%s'"
		return nil, bosherr.Errorf(msg, resp.StatusCode, respBody)
	}

	return respBody, nil
}

func (r HTTPClient) namespacedName(name string) string {
	return filepath.Join(r.namespace, name)
}

func (c HTTPClient) namespaceCAIfNecessary(i interface{}) interface{} {
	switch x := i.(type) {
	case map[string]interface{}:
		if caVal, found := x["ca"]; found {
			if caStr, ok := caVal.(string); ok {
				x["ca"] = c.namespacedName(caStr)
			}
		}
	}
	return i
}

func (c HTTPClient) stripInterfaceKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = c.stripInterfaceKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = c.stripInterfaceKeys(v)
		}
	}
	return i
}

func (c HTTPClient) putInterfaceKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[string]interface{}:
		m2 := map[interface{}]interface{}{}
		for k, v := range x {
			m2[k] = c.putInterfaceKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = c.putInterfaceKeys(v)
		}
	}
	return i
}
