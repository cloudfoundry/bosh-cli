package fsext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type InMemoryFS struct {
	contents map[string][]byte
}

var _ boshsys.FileSystem = &InMemoryFS{}

func NewInMemoryFS() *InMemoryFS {
	return &InMemoryFS{map[string][]byte{}}
}

func (f *InMemoryFS) ExpandPath(path string) (string, error) {
	return path, nil
}

func (f *InMemoryFS) ReadFileString(path string) (string, error) {
	bytes, err := f.ReadFile(path)
	return string(bytes), err
}

func (f *InMemoryFS) ReadFile(path string) ([]byte, error) {
	return f.ReadFileWithOpts(path, boshsys.ReadOpts{})
}

func (f *InMemoryFS) ReadFileWithOpts(path string, opts boshsys.ReadOpts) ([]byte, error) {
	if contents, found := f.contents[path]; found {
		return contents, nil
	}
	return nil, fmt.Errorf("Expected to find '%s'", path)
}

func (f *InMemoryFS) FileExists(path string) bool {
	_, found := f.contents[path]
	return found
}

func (f *InMemoryFS) WriteFileString(path, content string) error {
	return f.WriteFile(path, []byte(content))
}

func (f *InMemoryFS) WriteFile(path string, content []byte) error {
	f.contents[path] = content
	return nil
}

func (f *InMemoryFS) RemoveAll(path string) error {
	var dirPath = path
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	for k, _ := range f.contents {
		if strings.HasPrefix(k, dirPath) || k == path {
			delete(f.contents, k)
		}
	}
	return nil
}

func (f *InMemoryFS) panicNotSupported() {
	panic("Unexpected call to fsext.InMemoryFS")
}

// No interesting implementations below

func (f *InMemoryFS) ChangeTempRoot(path string) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) HomeDir(username string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *InMemoryFS) MkdirAll(path string, perm os.FileMode) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) Chown(path, username string) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) Chmod(path string, perm os.FileMode) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) OpenFile(path string, flag int, perm os.FileMode) (boshsys.File, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) ConvergeFileContents(path string, content []byte, opts ...boshsys.ConvergeFileContentsOpts) (bool, error) {
	f.panicNotSupported()
	return false, nil
}

func (f *InMemoryFS) Rename(oldPath, newPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) Symlink(oldPath, newPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) ReadAndFollowLink(symlinkPath string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *InMemoryFS) Readlink(symlinkPath string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *InMemoryFS) CopyFile(srcPath, dstPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) CopyDir(srcPath, dstPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) TempFile(prefix string) (boshsys.File, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) TempDir(prefix string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *InMemoryFS) Lstat(path string) (os.FileInfo, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) Stat(path string) (os.FileInfo, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) StatWithOpts(path string, opts boshsys.StatOpts) (os.FileInfo, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) RecursiveGlob(pattern string) ([]string, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) WriteFileQuietly(path string, content []byte) error {
	f.panicNotSupported()
	return nil
}

func (f *InMemoryFS) Glob(pattern string) ([]string, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *InMemoryFS) Walk(root string, walkFunc filepath.WalkFunc) error {
	f.panicNotSupported()
	return nil
}
