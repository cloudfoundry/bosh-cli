package releasedir

import (
	gobytes "bytes"
	//	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type DavConfig struct {
	User          string
	Password      string
	Endpoint      string
	Artifactory   bool
	RetryAttempts uint
}

func newClient(config DavConfig, httpClient httpclient.Client, digestCalculator bicrypto.DigestCalculator, logger boshlog.Logger) (*client, error) {
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

	cli := &client{config: config, httpClient: retryClient, digestCalculator: digestCalculator}
	return cli, nil
}

type client struct {
	config           DavConfig
	httpClient       httpclient.Client
	digestCalculator bicrypto.DigestCalculator
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

func (c client) Put(content io.ReadSeeker, path string, local_path string) (err error) {

	_, err = content.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("finding buffer position: %v", err)
	}
	// if artifactory
	if c.config.Artifactory {
		// Add artifactory properties
		// filepath
		res := strings.Split(local_path, "/blobs/")
		relative_path := res[1]
		path = path + ";filepath=" + relative_path
	}

	req, err := c.createReq("PUT", path, content)
	if err != nil {
		return err
	}

	// if artifactory
	if c.config.Artifactory {
		// TODO We need to take it from the blob.yml
		// sha1
		sha1, _ := c.digestCalculator.Calculate(local_path)
		if err != nil {
			return fmt.Errorf("Error calculating digest for file %s: %s", local_path, err)
		}
		// End TODO
		req.Header.Add("X-Checksum-Sha1", sha1)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return bosherr.WrapErrorf(err, "Putting dav blob %s", path)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Putting dav blob %s: Wrong response code: %d; body: %s", path, resp.StatusCode, c.readAndTruncateBody(resp))
	}

	return nil
}

func (c client) createReq(method, blobID string, body io.Reader) (*http.Request, error) {
	blobURL, err := url.Parse(c.config.Endpoint)
	if err != nil {
		return nil, err
	}

	/*
	       TODO do we want to use that ?
	   	digester := sha1.New()
	   	digester.Write([]byte(blobID))
	   	blobPrefix := fmt.Sprintf("%02x", digester.Sum(nil)[0])

	   	newPath := path.Join(blobURL.Path, blobPrefix, blobID)
	   	if !strings.HasPrefix(newPath, "/") {
	   		newPath = "/" + newPath
	   	}

	   	blobURL.Path = newPath

	       Or just that ???
	*/
	blobURL.Path = path.Join(blobURL.Path, blobID)

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

type DavBlobstore struct {
	fs      boshsys.FileSystem
	uuidGen boshuuid.Generator
	options map[string]interface{}
}

func NewDavBlobstore(
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	options map[string]interface{},
) DavBlobstore {
	return DavBlobstore{
		fs:      fs,
		uuidGen: uuidGen,
		options: options,
	}
}

func (b DavBlobstore) Get(blobID string) (string, error) {
	client, err := b.client()
	if err != nil {
		return "", err
	}

	file, err := b.fs.TempFile("bosh-dav-blob")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating destination file")
	}
	defer file.Close()

	//if err := client.Get(blobID, file); err != nil {
	reader, err := client.Get(blobID)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(file, reader)

	return file.Name(), nil
}

func (b DavBlobstore) Create(path string) (string, error) {
	client, err := b.client()
	if err != nil {
		return "", err
	}

	blobID, err := b.uuidGen.Generate()
	if err != nil {
		return "", bosherr.WrapError(err, "Generating blobstore ID")
	}

	file, err := b.fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", bosherr.WrapError(err, "Opening source file")
	}
	defer file.Close()

	if err := client.Put(file, blobID, path); err != nil {
		return "", err
	}

	return blobID, nil
}

func (b DavBlobstore) CleanUp(path string) error {
	return b.fs.RemoveAll(path)
}

func (b DavBlobstore) Delete(blobID string) error {
	panic("Not implemented")
}

func (b DavBlobstore) Validate() error {
	_, err := b.client()
	return err
}

func (b DavBlobstore) client() (*client, error) {

	conf := DavConfig{}
	bytes, err := json.Marshal(b.options)
	if err != nil {
		return nil, bosherr.WrapError(err, "Marshaling config")
	}
	bytes, err = ioutil.ReadAll(gobytes.NewBuffer(bytes))
	if err != nil {
		return &client{}, err
	}
	err = json.Unmarshal(bytes, &conf)
	if err != nil {
		return &client{}, err
	}

	// TODO We neeed to take it from the blob.yml
	digestCalculator := bicrypto.NewDigestCalculator(b.fs, []boshcrypto.Algorithm{boshcrypto.DigestAlgorithmSHA1})
	// End TODO
	logger := boshlog.NewLogger(boshlog.LevelNone)
	webdavclient, nil := newClient(conf, httpclient.DefaultClient, digestCalculator, logger)

	return webdavclient, nil
}
