package installation

import (
	"path/filepath"
)

type Target struct {
	path       string
	packageDir string
}

func NewTarget(path string, packageDir string) Target {
	return Target{
		path,
		packageDir,
	}
}

func (t Target) Path() string {
	return t.path
}

func (t Target) BlobstorePath() string {
	return filepath.Join(t.path, "blobs")
}

func (t Target) CompiledPackagedIndexPath() string {
	return filepath.Join(t.path, "compiled_packages.json")
}

func (t Target) TemplatesIndexPath() string {
	return filepath.Join(t.path, "templates.json")
}

func (t Target) PackagesPath() string {
	if t.packageDir != "" {
		return t.packageDir
	} else {
		return filepath.Join(t.path, "packages")
	}
}

func (t Target) JobsPath() string {
	return filepath.Join(t.path, "jobs")
}

func (t Target) TmpPath() string {
	return filepath.Join(t.path, "tmp")
}
