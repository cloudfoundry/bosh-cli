package util

import (
	"io"
	"os"
	"path/filepath"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

// WritePrivateFile restricts access before writing file contents.
func WritePrivateFile(fs boshsys.FileSystem, filename string, contents []byte) error {
	if err := fs.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}
	file, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck
	if err := fs.Chmod(filename, 0600); err != nil {
		return err
	}
	n, err := file.Write(contents)
	if err == nil && n != len(contents) {
		return io.ErrShortWrite
	}
	return err
}
