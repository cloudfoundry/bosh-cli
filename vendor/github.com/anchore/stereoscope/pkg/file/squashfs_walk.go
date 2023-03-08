package file

import (
	"io"
	"io/fs"

	"github.com/sylabs/squashfs"
)

// SquashFSVisitor is the type of the function called by WalkSquashFS to visit each file or
// directory.
//
// The path argument contains the full path, and d is the corresponding fs.DirEntry.
//
// The error result returned by the function controls how WalkSquashFS continues. If the function
// returns the special value fs.SkipDir, WalkSquashFS skips the current directory (path if
// d.IsDir() is true, otherwise path's parent directory). Otherwise, if the function returns a non-
// nil error, WalkSquashFS stops entirely and returns that error.
type SquashFSVisitor func(fsys fs.FS, path string, d fs.DirEntry) error

// WalkSquashFS walks the file tree within the SquashFS filesystem read from r, calling fn for each
// file or directory in the tree, including root.
func WalkSquashFS(r io.ReaderAt, fn SquashFSVisitor) error {
	fsys, err := squashfs.NewReader(r)
	if err != nil {
		return err
	}

	return fs.WalkDir(fsys, ".", walkDir(fsys, fn))
}

// WalkSquashFSFromReader walks the file tree within the SquashFS filesystem read from r, calling
// fn for each file or directory in the tree, including root. Callers should use WalkSquashFS
// where possible, as this function is considerably less efficient.
func WalkSquashFSFromReader(r io.Reader, fn SquashFSVisitor) error {
	fsys, err := squashfs.NewReaderFromReader(r)
	if err != nil {
		return err
	}

	return fs.WalkDir(fsys, ".", walkDir(fsys, fn))
}

// walkDir returns a fs.WalkDirFunc bound to fn.
func walkDir(fsys fs.FS, fn SquashFSVisitor) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		return fn(fsys, path, d)
	}
}
