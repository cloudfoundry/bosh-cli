package completion

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/peterbourgon/diskv"
)

type CacheItem struct {
	Timestamp time.Time `json:"timestamp"`
	Values    []string  `json:"values"`
}

type CompleteCache struct {
	store            *diskv.Diskv
	cacheKey         string
	expirationPeriod time.Duration
	logger           boshlog.Logger
	logTag           string
}

func NewCompleteCache(logger boshlog.Logger, cmdContext *CmdContext, group string) *CompleteCache {
	c := &CompleteCache{
		expirationPeriod: 15 * time.Second,
		logger:           logger,
		logTag:           "completion.CompleteCache",
	}
	c.configure(cmdContext, group)
	return c
}
func (c *CompleteCache) String() string {
	return c.store.BasePath + " / " + c.cacheKey
}

func (c *CompleteCache) configure(cmdContext *CmdContext, group string) {
	c.store = diskv.New(diskv.Options{
		BasePath: c.prepareCacheDir(cmdContext),
	})
	env := c.normalizeString(cmdContext.EnvironmentName)
	if env == "" {
		env = "default"
	}
	cacheKey := "env-" + env
	if cmdContext.DeploymentName != "" {
		cacheKey = cacheKey + "." + cmdContext.DeploymentName
	}
	if group != "" {
		c.cacheKey = cacheKey + "." + group
	}
}

func (c *CompleteCache) normalizeString(str string) string {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", str)
	if matched {
		hash := md5.Sum([]byte(str)) //nolint:gosec
		return hex.EncodeToString(hash[:])
	} else {
		return str
	}
}
func (c *CompleteCache) prepareCacheDir(cmdContext *CmdContext) string {
	basePath, _ := c.normalizeDirPath(cmdContext.ConfigPath)
	basePath = filepath.Join(filepath.Dir(basePath), "completion-cache")
	c.logger.Debug(c.logTag, "cache directory '%s', config directory '%s'", basePath, cmdContext.ConfigPath)
	if err := c.ensureDirExists(basePath); err != nil {
		c.logger.Debug(c.logTag, "creating cache directory '%s' error: %v", basePath, err)
		return ""
	}
	c.logger.Debug(c.logTag, "cache store: '%s'", basePath)
	return basePath
}

func (c *CompleteCache) normalizeDirPath(path string) (string, error) {
	sep := string(filepath.Separator)
	path = strings.TrimSuffix(path, string(filepath.Separator))
	if strings.HasPrefix(path, "~"+sep) {
		home, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		path = home + sep + path[len(sep)+1:]
	}
	return filepath.Abs(path)
}

func (c *CompleteCache) ensureDirExists(dirPath string) error {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not directory '%s'", dirPath)
	}
	return nil
}

func (c *CompleteCache) PutValues(values []string) error {
	item := CacheItem{
		Timestamp: time.Now(),
		Values:    values,
	}
	data, err := json.Marshal(item)
	if err != nil {
		c.logger.Debug(c.logTag, "marshal values error: %v", err)
		return err
	}
	err = c.store.Write(c.cacheKey, data)
	if err != nil {
		c.logger.Debug(c.logTag, "write cache error: %v", err)
	}
	return err
}

func (c *CompleteCache) GetValues() (values []string, valid bool, err error) {
	data, err := c.store.Read(c.cacheKey)
	if err != nil {
		if err != nil {
			c.logger.Debug(c.logTag, "read cache error: %v", err)
		}
		return nil, false, err
	}

	var item CacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		if err != nil {
			c.logger.Debug(c.logTag, "unmarshal cached values error: %v", err)
		}
		return nil, false, err
	}
	valid = time.Since(item.Timestamp) < c.expirationPeriod
	if !valid {
		_ = c.store.Erase(c.cacheKey)
	}
	return item.Values, valid, nil
}
