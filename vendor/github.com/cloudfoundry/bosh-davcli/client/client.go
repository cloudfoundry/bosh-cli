package client

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	davconf "github.com/cloudfoundry/bosh-davcli/config"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Client interface {
	Get(path string) (content io.ReadCloser, err error)
	Put(path string, content io.ReadCloser, contentLength int64) (err error)
	Exists(path string) (err error)
	Delete(path string) (err error)
}

func NewClient(config davconf.Config, httpClient httpclient.Client, logger boshlog.Logger) (c Client) {
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	// @todo should a logger now be passed in to this client?
	duration := time.Duration(0)
	retryClient := httpclient.NewRetryClient(
		httpClient,
		config.RetryAttempts,
		duration,
		logger,
	)

	return client{
		config:     config,
		httpClient: retryClient,
	}
}

type client struct {
	config     davconf.Config
	httpClient httpclient.Client
}

func (c client) Get(path string) (io.ReadCloser, error) {
	req, err := c.createReq("GET", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting dav blob %s", path)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Getting dav blob %s: Wrong response code: %d; body: %s", path, resp.StatusCode, c.readAndTruncateBody(resp))
	}

	return resp.Body, nil
}

func (c client) Put(path string, content io.ReadCloser, contentLength int64) error {
	req, err := c.createReq("PUT", path, content)
	if err != nil {
		return err
	}
	defer content.Close()

	req.ContentLength = contentLength
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return bosherr.WrapErrorf(err, "Putting dav blob %s", path)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Putting dav blob %s: Wrong response code: %d; body: %s", path, resp.StatusCode, c.readAndTruncateBody(resp))
	}

	return nil
}

func (c client) Exists(path string) error {
	req, err := c.createReq("HEAD", path, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking if dav blob %s exists", path)
	}

	if resp.StatusCode == http.StatusNotFound {
		err := fmt.Errorf("%s not found", path)
		return bosherr.WrapErrorf(err, "Checking if dav blob %s exists", path)
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("invalid status: %d", resp.StatusCode)
		return bosherr.WrapErrorf(err, "Checking if dav blob %s exists", path)
	}

	return nil
}

func (c client) Delete(path string) error {
	req, err := c.createReq("DELETE", path, nil)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating delete request for blob '%s'", path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting blob '%s'", path)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err := fmt.Errorf("invalid status: %d", resp.StatusCode)
		return bosherr.WrapErrorf(err, "Deleting blob '%s'", path)
	}

	return nil
}

func (c client) createReq(method, blobID string, body io.Reader) (*http.Request, error) {
	blobURL, err := url.Parse(c.config.Endpoint)
	if err != nil {
		return nil, err
	}

	digester := sha1.New()
	digester.Write([]byte(blobID))
	blobPrefix := fmt.Sprintf("%02x", digester.Sum(nil)[0])

	newPath := path.Join(blobURL.Path, blobPrefix, blobID)
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/" + newPath
	}

	blobURL.Path = newPath

	req, err := http.NewRequest(method, blobURL.String(), body)
	if err != nil {
		return req, err
	}

	req.SetBasicAuth(c.config.User, c.config.Password)
	return req, nil
}

func (c client) readAndTruncateBody(resp *http.Response) string {
	body := ""
	if resp.Body != nil {
		buf := make([]byte, 1024)
		n, err := resp.Body.Read(buf)
		if err == io.EOF || err == nil {
			body = string(buf[0:n])
		}
	}
	return body
}
