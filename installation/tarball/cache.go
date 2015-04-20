package tarball

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type Cache interface {
	Get(sha1 string) (path string, found bool)
	Save(sourcePath string, sha1 string) (path string, err error)
}

type cache struct {
	basePath string
	fs       boshsys.FileSystem
	logger   boshlog.Logger
	logTag   string
}

func NewCache(basePath string, fs boshsys.FileSystem, logger boshlog.Logger) Cache {
	return &cache{
		basePath: basePath,
		fs:       fs,
		logger:   logger,
		logTag:   "tarballCache",
	}
}

func (c *cache) Get(sha1 string) (string, bool) {
	path := filepath.Join(c.basePath, sha1)
	if c.fs.FileExists(path) {
		return path, true
	}

	return "", false
}

func (c *cache) Save(sourcePath string, sha1 string) (string, error) {
	err := c.fs.MkdirAll(c.basePath, os.FileMode(0766))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to create cache directory '%s'", c.basePath)
	}

	path := filepath.Join(c.basePath, sha1)

	err = c.fs.Rename(sourcePath, path)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Failed to save tarball path '%s' in cache", sourcePath)
	}

	return path, nil
}
