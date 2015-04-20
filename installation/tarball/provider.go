package tarball

import (
	"io"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bicrypto "github.com/cloudfoundry/bosh-init/crypto"
	bihttpclient "github.com/cloudfoundry/bosh-init/deployment/httpclient"
)

type Source interface {
	URL() string
	SHA1() string
}

type Provider interface {
	Get(Source) (path string, err error)
}

type provider struct {
	cache          Cache
	fs             boshsys.FileSystem
	httpClient     bihttpclient.HTTPClient
	sha1Calculator bicrypto.SHA1Calculator
	logger         boshlog.Logger
	logTag         string
}

func NewProvider(
	cache Cache,
	fs boshsys.FileSystem,
	httpClient bihttpclient.HTTPClient,
	sha1Calculator bicrypto.SHA1Calculator,
	logger boshlog.Logger,
) Provider {
	return &provider{
		cache:          cache,
		fs:             fs,
		httpClient:     httpClient,
		sha1Calculator: sha1Calculator,
		logger:         logger,
		logTag:         "tarballProvider",
	}
}

func (p *provider) Get(source Source) (string, error) {
	if strings.HasPrefix(source.URL(), "file://") {
		return strings.TrimPrefix(source.URL(), "file://"), nil
	}

	if !strings.HasPrefix(source.URL(), "http") {
		return "", bosherr.Errorf("Invalid source URL: '%s', must be either file:// or http(s)://", source.URL())
	}

	cachedPath, found := p.cache.Get(source.SHA1())
	if found {
		return cachedPath, nil
	}

	downloadedFile, err := p.fs.TempFile("tarballProvider")
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to create temporary file when downloading: '%s'", source.URL())
	}
	defer p.fs.RemoveAll(downloadedFile.Name())

	response, err := p.httpClient.Get(source.URL())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to download from endpoint: '%s'", source.URL())
	}
	defer response.Body.Close()

	_, err = io.Copy(downloadedFile, response.Body)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to download to temporary file from endpoint: '%s'", source.URL())
	}

	downloadedSha1, err := p.sha1Calculator.Calculate(downloadedFile.Name())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to calculate sha1 for downloaded file from endpoint: '%s'", source.URL())
	}

	if downloadedSha1 != source.SHA1() {
		return "", bosherr.Errorf("SHA1 of downloaded file '%s' does not match source SHA1 '%s'", downloadedSha1, source.SHA1())
	}

	cachedPath, err = p.cache.Save(downloadedFile.Name(), source.SHA1())
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to save tarball in cache from endpoint: '%s'", source.URL())
	}

	return cachedPath, nil
}
