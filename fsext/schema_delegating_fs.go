package fsext

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type SchemaDelegatingFS struct {
	defaultFS boshsys.FileSystem
	schemas   map[string]boshsys.FileSystem
}

var _ boshsys.FileSystem = &SchemaDelegatingFS{}

func NewSchemaDelegatingFS(defaultFS boshsys.FileSystem) *SchemaDelegatingFS {
	return &SchemaDelegatingFS{defaultFS, map[string]boshsys.FileSystem{}}
}

func (f *SchemaDelegatingFS) RegisterSchema(schema string, fs boshsys.FileSystem) {
	f.schemas[schema] = fs
}

func (f *SchemaDelegatingFS) fs(path string) boshsys.FileSystem {
	if !strings.Contains(path, "://") {
		return f.defaultFS
	}

	pieces := strings.SplitN(path, "://", 2)

	if fs, found := f.schemas[pieces[0]]; found {
		return fs
	}

	panic(fmt.Sprintf("Unknown schema '%s' in path '%s'", pieces[0], path))
}

func (f *SchemaDelegatingFS) stripSchema(path string) string {
	if strings.Contains(path, "://") {
		pieces := strings.SplitN(path, "://", 2)
		return pieces[1]
	}
	return path
}

func (f *SchemaDelegatingFS) addSchema(origPath, newPath string) string {
	if strings.Contains(origPath, "://") {
		pieces := strings.SplitN(origPath, "://", 2)
		return pieces[0] + "://" + newPath
	}
	return newPath
}

func (f *SchemaDelegatingFS) ExpandPath(path string) (string, error) {
	result, err := f.fs(path).ExpandPath(f.stripSchema(path))
	if err != nil {
		return "", err
	}
	return f.addSchema(path, result), nil
}

func (f *SchemaDelegatingFS) ReadFileString(path string) (string, error) {
	return f.fs(path).ReadFileString(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) ReadFile(path string) ([]byte, error) {
	return f.fs(path).ReadFile(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) ReadFileWithOpts(path string, opts boshsys.ReadOpts) ([]byte, error) {
	return f.fs(path).ReadFileWithOpts(f.stripSchema(path), opts)
}

func (f *SchemaDelegatingFS) FileExists(path string) bool {
	return f.fs(path).FileExists(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) WriteFileString(path, content string) error {
	return f.fs(path).WriteFileString(f.stripSchema(path), content)
}

func (f *SchemaDelegatingFS) WriteFile(path string, content []byte) error {
	return f.fs(path).WriteFile(f.stripSchema(path), content)
}

func (f *SchemaDelegatingFS) ChangeTempRoot(path string) error {
	return f.fs(path).ChangeTempRoot(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) HomeDir(username string) (string, error) {
	return f.defaultFS.HomeDir(username)
}

func (f *SchemaDelegatingFS) MkdirAll(path string, perm os.FileMode) error {
	return f.fs(path).MkdirAll(f.stripSchema(path), perm)
}

func (f *SchemaDelegatingFS) RemoveAll(path string) error {
	return f.fs(path).RemoveAll(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) Chown(path, username string) error {
	return f.fs(path).Chown(f.stripSchema(path), username)
}

func (f *SchemaDelegatingFS) Chmod(path string, perm os.FileMode) error {
	return f.fs(path).Chmod(f.stripSchema(path), perm)
}

func (f *SchemaDelegatingFS) OpenFile(path string, flag int, perm os.FileMode) (boshsys.File, error) {
	return f.fs(path).OpenFile(f.stripSchema(path), flag, perm)
}

func (f *SchemaDelegatingFS) ConvergeFileContents(path string, content []byte, opts ...boshsys.ConvergeFileContentsOpts) (bool, error) {
	return f.fs(path).ConvergeFileContents(f.stripSchema(path), content, opts...)
}

func (f *SchemaDelegatingFS) Rename(oldPath, newPath string) error {
	return f.fs(oldPath).Rename(f.stripSchema(oldPath), f.stripSchema(newPath))
}

func (f *SchemaDelegatingFS) Symlink(oldPath, newPath string) error {
	return f.fs(oldPath).Symlink(f.stripSchema(oldPath), f.stripSchema(newPath))
}

func (f *SchemaDelegatingFS) ReadAndFollowLink(path string) (string, error) {
	return f.fs(path).ReadAndFollowLink(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) Readlink(path string) (string, error) {
	return f.fs(path).Readlink(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) CopyFile(srcPath, dstPath string) error {
	return f.fs(srcPath).CopyFile(f.stripSchema(srcPath), f.stripSchema(dstPath))
}

func (f *SchemaDelegatingFS) CopyDir(srcPath, dstPath string) error {
	return f.fs(srcPath).CopyDir(f.stripSchema(srcPath), f.stripSchema(dstPath))
}

func (f *SchemaDelegatingFS) TempFile(prefix string) (boshsys.File, error) {
	return f.defaultFS.TempFile(prefix)
}

func (f *SchemaDelegatingFS) TempDir(prefix string) (string, error) {
	return f.defaultFS.TempDir(prefix)
}

func (f *SchemaDelegatingFS) Lstat(path string) (os.FileInfo, error) {
	return f.fs(path).Lstat(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) Stat(path string) (os.FileInfo, error) {
	return f.fs(path).Stat(f.stripSchema(path))
}

func (f *SchemaDelegatingFS) StatWithOpts(path string, opts boshsys.StatOpts) (os.FileInfo, error) {
	return f.fs(path).StatWithOpts(f.stripSchema(path), opts)
}

func (f *SchemaDelegatingFS) RecursiveGlob(pattern string) ([]string, error) {
	return f.fs(pattern).RecursiveGlob(f.stripSchema(pattern))
}

func (f *SchemaDelegatingFS) WriteFileQuietly(path string, content []byte) error {
	return f.fs(path).WriteFileQuietly(f.stripSchema(path), content)
}

func (f *SchemaDelegatingFS) Glob(pattern string) ([]string, error) {
	return f.fs(pattern).Glob(f.stripSchema(pattern))
}

func (f *SchemaDelegatingFS) Walk(root string, walkFunc filepath.WalkFunc) error {
	return f.fs(root).Walk(f.stripSchema(root), walkFunc)
}
