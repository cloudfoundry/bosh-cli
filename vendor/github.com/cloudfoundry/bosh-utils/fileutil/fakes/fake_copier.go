package fakes

import "github.com/cloudfoundry/bosh-utils/fileutil"

type FakeCopier struct {
	FilteredCopyToTempTempDir      string
	FilteredCopyToTempError        error
	FilteredCopyToTempDir          string
	FilteredCopyToTempFilters      []string
	FilteredMultiCopyToTempDirs    []fileutil.DirToCopy
	FilteredMultiCopyToTempFilters []string
	FilteredMultiCopyToTempError   error
	FilteredMultiCopyToTempDir     string

	CleanUpTempDir string
}

func NewFakeCopier() (copier *FakeCopier) {
	copier = &FakeCopier{}
	return
}

func (c *FakeCopier) FilteredCopyToTemp(dir string, filters []string) (tempDir string, err error) {
	c.FilteredCopyToTempDir = dir
	c.FilteredCopyToTempFilters = filters
	tempDir = c.FilteredCopyToTempTempDir
	err = c.FilteredCopyToTempError
	return
}

func (c *FakeCopier) FilteredMultiCopyToTemp(dirs []fileutil.DirToCopy, filters []string) (tempDir string, err error) {
	c.FilteredMultiCopyToTempDirs = dirs
	c.FilteredMultiCopyToTempFilters = filters
	tempDir = c.FilteredMultiCopyToTempDir
	err = c.FilteredMultiCopyToTempError
	return
}

func (c *FakeCopier) CleanUp(tempDir string) {
	c.CleanUpTempDir = tempDir
}
