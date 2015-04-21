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
	Path(sha1 string) (path string)
	Save(sourcePath string, sha1 string) error
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
	cachedPath := c.Path(sha1)
	if c.fs.FileExists(cachedPath) {
		c.logger.Debug(c.logTag, "Found cached tarball at: '%s'", cachedPath)
		return cachedPath, true
	}

	return "", false
}

func (c *cache) Save(sourcePath string, sha1 string) error {
	err := c.fs.MkdirAll(c.basePath, os.FileMode(0766))
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to create cache directory '%s'", c.basePath)
	}

	err = c.fs.Rename(sourcePath, c.Path(sha1))
	if err != nil {
		return bosherr.WrapErrorf(err, "Failed to save tarball path '%s' in cache", sourcePath)
	}

	c.logger.Debug(c.logTag, "Saving tarball in cache at: '%s'", c.Path(sha1))
	return nil
}

func (c *cache) Path(sha1 string) string {
	return filepath.Join(c.basePath, sha1)
}
