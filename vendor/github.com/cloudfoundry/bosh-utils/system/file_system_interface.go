package system

import (
	"io"
	"os"
	"path/filepath"
)

// File is a subset of os.File
type File interface {
	io.ReadWriteCloser
	ReadAt([]byte, int64) (int, error)
	WriteAt([]byte, int64) (int, error)
	Seek(int64, int) (int64, error)
	Stat() (os.FileInfo, error)
	Name() string
}

type FileSystem interface {
	HomeDir(username string) (path string, err error)
	ExpandPath(path string) (expandedPath string, err error)

	// MkdirAll will not change existing dir permissions
	// if dir exists and has different permissions
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(fileOrDir string) error

	Chown(path, username string) error
	Chmod(path string, perm os.FileMode) error

	OpenFile(path string, flag int, perm os.FileMode) (File, error)

	WriteFileString(path, content string) error
	WriteFile(path string, content []byte) error
	WriteFileQuietly(path string, content []byte) error
	ConvergeFileContents(path string, content []byte, opts ...ConvergeFileContentsOpts) (written bool, err error)

	ReadFileString(path string) (content string, err error)
	ReadFile(path string) (content []byte, err error)
	ReadFileWithOpts(path string, opts ReadOpts) (content []byte, err error)

	FileExists(path string) bool
	Stat(path string) (os.FileInfo, error)
	StatWithOpts(path string, opts StatOpts) (os.FileInfo, error)
	Lstat(path string) (os.FileInfo, error)

	Rename(oldPath, newPath string) error

	// Symlink creates a symlink file at newPath pointing to file at oldPath,
	// and removes any file that might exist at newPath.
	Symlink(oldPath, newPath string) error

	ReadAndFollowLink(symlinkPath string) (targetPath string, err error)
	Readlink(symlinkPath string) (targetPath string, err error)

	CopyFile(srcPath, dstPath string) error
	CopyDir(srcPath, dstPath string) error

	// TempFile returns a unique temporary file with a custom prefix
	TempFile(prefix string) (file File, err error)
	TempDir(prefix string) (path string, err error)
	ChangeTempRoot(path string) error

	Glob(pattern string) (matches []string, err error)
	RecursiveGlob(pattern string) (matches []string, err error)
	Walk(root string, walkFunc filepath.WalkFunc) error
}
