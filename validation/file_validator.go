package validation

import (
	"fmt"

	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type FileValidator interface {
	Exists(path string) error
}

type fileValidator struct {
	fs boshsys.FileSystem
}

func NewFileValidator(fs boshsys.FileSystem) fileValidator {
	return fileValidator{fs: fs}
}

func (v fileValidator) Exists(path string) error {
	if !v.fs.FileExists(path) {
		return fmt.Errorf("Path '%s' does not exist", path)
	}
	return nil
}
