package configserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type FS struct {
	client Client
}

var _ boshsys.FileSystem = &FS{}

func NewFS(client Client) *FS {
	return &FS{client}
}

func (f *FS) ExpandPath(path string) (string, error) {
	return path, nil
}

func (f *FS) ReadFileString(path string) (string, error) {
	bytes, err := f.ReadFile(path)
	return string(bytes), err
}

func (f *FS) ReadFile(path string) ([]byte, error) {
	return f.ReadFileWithOpts(path, boshsys.ReadOpts{})
}

func (f *FS) ReadFileWithOpts(path string, opts boshsys.ReadOpts) ([]byte, error) {
	res, err := f.client.Read(f.nameForPath(path))
	if err != nil {
		return nil, err
	}

	return []byte(res.(string)), nil
}

func (f *FS) FileExists(path string) bool {
	found, err := f.client.Exists(f.nameForPath(path))
	if err != nil {
		panic(fmt.Sprintf("Unexpected config server error: %s", err))
	}

	return found
}

func (f *FS) WriteFileString(path, content string) error {
	return f.WriteFile(path, []byte(content))
}

func (f *FS) WriteFile(path string, content []byte) error {
	return f.client.Write(f.nameForPath(path), content)
}

func (f *FS) RemoveAll(path string) error {
	return f.client.Delete(f.nameForPath(path))
}

func (f *FS) nameForPath(path string) string {
	return strings.Replace(path, ".", "-", -1)
}

func (f *FS) panicNotSupported() {
	panic("Unexpected call to configserver.FS")
}

// No interesting implementations below

func (f *FS) ChangeTempRoot(path string) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) HomeDir(username string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *FS) MkdirAll(path string, perm os.FileMode) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) Chown(path, username string) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) Chmod(path string, perm os.FileMode) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) OpenFile(path string, flag int, perm os.FileMode) (boshsys.File, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) ConvergeFileContents(path string, content []byte, opts ...boshsys.ConvergeFileContentsOpts) (bool, error) {
	f.panicNotSupported()
	return false, nil
}

func (f *FS) Rename(oldPath, newPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) Symlink(oldPath, newPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) ReadAndFollowLink(symlinkPath string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *FS) Readlink(symlinkPath string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *FS) CopyFile(srcPath, dstPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) CopyDir(srcPath, dstPath string) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) TempFile(prefix string) (boshsys.File, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) TempDir(prefix string) (string, error) {
	f.panicNotSupported()
	return "", nil
}

func (f *FS) Lstat(path string) (os.FileInfo, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) Stat(path string) (os.FileInfo, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) StatWithOpts(path string, opts boshsys.StatOpts) (os.FileInfo, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) RecursiveGlob(pattern string) ([]string, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) WriteFileQuietly(path string, content []byte) error {
	f.panicNotSupported()
	return nil
}

func (f *FS) Glob(pattern string) ([]string, error) {
	f.panicNotSupported()
	return nil, nil
}

func (f *FS) Walk(root string, walkFunc filepath.WalkFunc) error {
	f.panicNotSupported()
	return nil
}
