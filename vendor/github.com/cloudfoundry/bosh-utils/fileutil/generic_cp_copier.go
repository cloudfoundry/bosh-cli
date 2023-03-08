package fileutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const genericCpCopierLogTag = "genericCpCopier"

type DirToCopy struct {
	Dir    string
	Prefix string
}

type genericCpCopier struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewGenericCpCopier(
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) Copier {
	return genericCpCopier{fs: fs, logger: logger}
}

func (c genericCpCopier) FilteredCopyToTemp(dir string, filters []string) (string, error) {
	return c.FilteredMultiCopyToTemp([]DirToCopy{{Dir: dir}}, filters)
}

func (c genericCpCopier) FilteredMultiCopyToTemp(dirs []DirToCopy, filters []string) (string, error) {
	var err error

	tempDir, err := c.fs.TempDir("bosh-platform-commands-cpCopier-FilteredCopyToTemp")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating temporary directory")
	}

	err = os.Chmod(tempDir, os.FileMode(0755))
	if err != nil {
		c.CleanUp(tempDir)
		bosherr.WrapError(err, "Fixing permissions on temp dir")
	}

	for _, dirToCopy := range dirs {
		globsFiles := c.convertDirectoriesToGlobs(dirToCopy.Dir, filters)
		var filesToCopy []string

		for _, globFile := range globsFiles {
			filteredFilesToCopy, err := doublestar.Glob(globFile)
			if err != nil {
				c.CleanUp(tempDir)
				return "", bosherr.WrapError(err, "Finding files matching filters")
			}

			for _, fileToCopy := range filteredFilesToCopy {
				filesToCopy = append(filesToCopy, strings.TrimPrefix(strings.TrimPrefix(fileToCopy, dirToCopy.Dir), "/"))
			}
		}

		err = c.copyFilesToDir(filesToCopy, dirToCopy.Dir, tempDir, dirToCopy.Prefix)
		if err != nil {
			c.CleanUp(tempDir)
			return "", bosherr.WrapError(err, "Copying Files to Temp Dir")
		}
	}

	return tempDir, nil
}

func (c genericCpCopier) CleanUp(tempDir string) {
	err := c.fs.RemoveAll(tempDir)
	if err != nil {
		c.logger.Error(genericCpCopierLogTag, "Failed to clean up temporary directory %s: %#v", tempDir, err)
	}
}

func (c genericCpCopier) copyFilesToDir(fileList []string, srcDir string, destDir string, destPrefix string) error {
	destDir = filepath.Join(destDir, destPrefix)

	for _, relativePath := range fileList {
		src := filepath.Join(srcDir, relativePath)
		dst := filepath.Join(destDir, relativePath)

		fileInfo, err := os.Stat(src)
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting file info for '%s'", src)
		}

		if !fileInfo.IsDir() {
			dstContainingDir := filepath.Dir(dst)
			err := c.fs.MkdirAll(dstContainingDir, os.ModePerm)
			if err != nil {
				return bosherr.WrapErrorf(err, "Making destination directory '%s' for '%s'", dstContainingDir, src)
			}

			err = c.fs.CopyFile(src, dst)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c genericCpCopier) convertDirectoriesToGlobs(dir string, filters []string) []string {
	convertedFilters := []string{}
	for _, filter := range filters {
		src := filepath.Join(dir, filter)
		fileInfo, err := os.Stat(src)
		if err == nil && fileInfo.IsDir() {
			convertedFilters = append(convertedFilters, filepath.Join(src, "**", "*"))
		} else {
			convertedFilters = append(convertedFilters, src)
		}
	}

	return convertedFilters
}
