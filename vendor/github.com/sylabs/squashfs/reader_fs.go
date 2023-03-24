package squashfs

import (
	"bytes"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/sylabs/squashfs/internal/directory"
	"github.com/sylabs/squashfs/internal/inode"
)

//FS is a fs.FS representation of a squashfs directory.
//Implements fs.GlobFS, fs.ReadDirFS, fs.ReadFileFS, fs.StatFS, and fs.SubFS
type FS struct {
	*File
	e []directory.Entry
}

func (r Reader) newFS(e directory.Entry, parent *FS) (*FS, error) {
	i, err := r.inodeFromDir(e)
	if err != nil {
		return nil, err
	}
	ents, err := r.readDirectory(i)
	if err != nil {
		return nil, err
	}
	return &FS{
		File: &File{
			i:      i,
			r:      &r,
			parent: parent,
			e:      e,
		},
		e: ents,
	}, nil
}

//Open opens the file at name. Returns a squashfs.File.
func (f FS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = filepath.Clean(name)
	if name == "." || name == "" {
		return f.File, nil
	}
	split := strings.Split(name, "/")
	for i := range f.e {
		if f.e[i].Name != split[0] {
			continue
		}
		if len(split) > 1 && f.e[i].Type != inode.Dir {
			return nil, &fs.PathError{
				Op:   "open",
				Path: name,
				Err:  fs.ErrNotExist,
			}
		}
		if len(split) > 1 {
			newFS, err := f.r.newFS(f.e[i], &f)
			if err != nil {
				return nil, &fs.PathError{
					Op:   "open",
					Path: name,
					Err:  err,
				}
			}
			out, err := newFS.Open(strings.Join(split[1:], "/"))
			if err != nil {
				err.(*fs.PathError).Path = name
			}
			return out, err
		}
		out, err := f.r.newFile(f.e[i], &f)
		if err != nil {
			err = &fs.PathError{
				Op:   "open",
				Path: name,
				Err:  err,
			}
		}
		return out, err
	}
	return nil, &fs.PathError{
		Op:   "open",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

//Glob returns the name of the files at the given pattern.
//All paths are relative to the FS.
//Uses filepath.Match to compare names.
func (f FS) Glob(pattern string) (out []string, err error) {
	if !fs.ValidPath(pattern) {
		return nil, &fs.PathError{
			Op:   "glob",
			Path: pattern,
			Err:  fs.ErrInvalid,
		}
	}
	pattern = filepath.Clean(pattern)
	split := strings.Split(pattern, "/")
	for i := 0; i < len(f.e); i++ {
		if match, _ := path.Match(split[0], f.e[i].Name); match {
			if len(split) == 1 {
				out = append(out, f.e[i].Name)
				continue
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "glob"
					pathErr.Path = pattern
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "glob",
					Path: pattern,
					Err:  err,
				}
			}
			subGlob, err := sub.(fs.GlobFS).Glob(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "glob"
					pathErr.Path = pattern
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "glob",
					Path: pattern,
					Err:  err,
				}
			}
			for i := 0; i < len(subGlob); i++ {
				subGlob[i] = f.File.e.Name + "/" + subGlob[i]
			}
			out = append(out, subGlob...)
		}
	}
	return
}

//ReadDir returns all the DirEntry returns all DirEntry's for the directory at name.
//If name is not a directory, returns an error.
func (f FS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "readdir",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = filepath.Clean(name)
	if name == "." || name == "" {
		return f.File.ReadDir(-1)
	}
	split := strings.Split(name, "/")
	for i := 0; i < len(f.e); i++ {
		if split[0] == f.e[i].Name {
			if len(split) == 1 {
				fi, err := f.r.newFile(f.e[i], &f)
				if err != nil {
					return nil, &fs.PathError{
						Op:   "readdir",
						Path: name,
						Err:  err,
					}
				}
				out, err := fi.ReadDir(-1)
				if err != nil {
					err = &fs.PathError{
						Op:   "readdir",
						Path: name,
						Err:  err,
					}
				}
				return out, err
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "readir"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "readdir",
					Path: name,
					Err:  err,
				}
			}
			redDir, err := sub.(fs.ReadDirFS).ReadDir(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "readdir"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "readdir",
					Path: name,
					Err:  err,
				}
			}
			return redDir, nil
		}
	}
	return nil, &fs.PathError{
		Op:   "readdir",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

//ReadFile returns the data (in []byte) for the file at name.
func (f FS) ReadFile(name string) ([]byte, error) {
	fil, err := f.Open(name)
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok {
			pathErr.Op = "readfile"
			pathErr.Path = name
			return nil, pathErr
		}
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, fil)
	if err != nil {
		return nil, &fs.PathError{
			Op:   "readfile",
			Path: name,
			Err:  err,
		}
	}
	return buf.Bytes(), nil
}

//Stat returns the fs.FileInfo for the file at name.
func (f FS) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "stat",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}
	name = filepath.Clean(strings.TrimPrefix(name, "/"))
	if name == "." || name == "" {
		return f.File.Stat()
	}
	split := strings.Split(name, "/")
	for i := 0; i < len(f.e); i++ {
		if split[0] == f.e[i].Name {
			if len(split) == 1 {
				in, err := f.r.newFileInfo(f.e[i])
				if err != nil {
					err = &fs.PathError{
						Op:   "stat",
						Path: name,
						Err:  err,
					}
				}
				return in, err
			}
			sub, err := f.Sub(split[0])
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "stat"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "stat",
					Path: name,
					Err:  err,
				}
			}
			stat, err := sub.(fs.StatFS).Stat(strings.Join(split[1:], "/"))
			if err != nil {
				if pathErr, ok := err.(*fs.PathError); ok {
					if pathErr.Err == fs.ErrNotExist {
						continue
					}
					pathErr.Op = "stat"
					pathErr.Path = name
					return nil, pathErr
				}
				return nil, &fs.PathError{
					Op:   "stat",
					Path: name,
					Err:  err,
				}
			}
			return stat, nil
		}
	}
	return nil, &fs.PathError{
		Op:   "stat",
		Path: name,
		Err:  fs.ErrNotExist,
	}
}

//Sub returns the FS at dir
func (f FS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, &fs.PathError{
			Op:   "sub",
			Path: dir,
			Err:  fs.ErrInvalid,
		}
	}
	dir = filepath.Clean(dir)
	if dir == "." || dir == "" {
		return f, nil
	}
	split := strings.Split(dir, "/")
	for i := range f.e {
		if f.e[i].Name != split[0] {
			continue
		}
		if f.e[i].Type != inode.Dir {
			return nil, &fs.PathError{
				Op:   "sub",
				Path: dir,
				Err:  fs.ErrNotExist,
			}
		}
		newFS, err := f.r.newFS(f.e[i], &f)
		if err != nil {
			return nil, &fs.PathError{
				Op:   "sub",
				Path: dir,
				Err:  err,
			}
		}
		if len(split) > 1 {
			ret, err := newFS.Sub(strings.Join(split[1:], "/"))
			if err != nil {
				err.(*fs.PathError).Path = dir
			}
			return ret, err
		}
		return newFS, nil
	}
	return nil, &fs.PathError{
		Op:   "sub",
		Path: dir,
		Err:  fs.ErrNotExist,
	}
}
