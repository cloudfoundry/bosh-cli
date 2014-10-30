package blobstore

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	boshdavcli "github.com/cloudfoundry/bosh-agent/davcli/client"
	boshdavcliconf "github.com/cloudfoundry/bosh-agent/davcli/config"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Factory interface {
	Create(string) (Blobstore, error)
}

type blobstoreFactory struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewBlobstoreFactory(fs boshsys.FileSystem, logger boshlog.Logger) Factory {
	return blobstoreFactory{
		fs:     fs,
		logger: logger,
	}
}

func (f blobstoreFactory) Create(blobstoreURL string) (Blobstore, error) {
	blobstoreConfig, err := f.parseBlobstoreURL(blobstoreURL)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating blobstore config")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := http.Client{Transport: tr}

	davClient := boshdavcli.NewClient(boshdavcliconf.Config{
		Endpoint: fmt.Sprintf("%s/blobs", blobstoreConfig.Endpoint),
		User:     blobstoreConfig.Username,
		Password: blobstoreConfig.Password,
	}, &httpClient)

	return NewBlobstore(davClient, f.fs, f.logger), nil
}

func (f blobstoreFactory) parseBlobstoreURL(blobstoreURL string) (Config, error) {
	parsedURL, err := url.Parse(blobstoreURL)
	if err != nil {
		return Config{}, bosherr.WrapError(err, "Parsing Mbus URL")
	}

	var username, password string
	userInfo := parsedURL.User
	if userInfo != nil {
		username = userInfo.Username()
		password, _ = userInfo.Password()
	}

	return Config{
		Endpoint: fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host),
		Username: username,
		Password: password,
	}, nil
}
