package squashfs

import (
	"io/fs"
	"time"

	"github.com/sylabs/squashfs/internal/directory"
	"github.com/sylabs/squashfs/internal/inode"
)

type fileInfo struct {
	e       directory.Entry
	size    int64
	perm    uint32
	modTime uint32
}

func (r Reader) newFileInfo(e directory.Entry) (fileInfo, error) {
	i, err := r.inodeFromDir(e)
	if err != nil {
		return fileInfo{}, err
	}
	return newFileInfo(e, i), nil
}

func newFileInfo(e directory.Entry, i inode.Inode) fileInfo {
	var size int64
	if i.Type == inode.Fil {
		size = int64(i.Data.(inode.File).Size)
	} else if i.Type == inode.EFil {
		size = int64(i.Data.(inode.EFile).Size)
	}
	return fileInfo{
		e:       e,
		size:    size,
		perm:    uint32(i.Perm),
		modTime: i.ModTime,
	}
}

func (f fileInfo) Name() string {
	return f.e.Name
}

func (f fileInfo) Size() int64 {
	return f.size
}

func (f fileInfo) Mode() fs.FileMode {
	if f.IsDir() {
		return fs.FileMode(f.perm | uint32(fs.ModeDir))
	}
	return fs.FileMode(f.perm)
}

func (f fileInfo) ModTime() time.Time {
	return time.Unix(int64(f.modTime), 0)
}

func (f fileInfo) IsDir() bool {
	return f.e.Type == inode.Dir
}

func (f fileInfo) Sys() any {
	return nil
}
