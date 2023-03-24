package file

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/anchore/stereoscope/internal/log"

	"github.com/sylabs/squashfs"
)

// Metadata represents all file metadata of interest (used today for in-tar file resolution).
type Metadata struct {
	// Path is the absolute path representation to the file
	Path string
	// LinkDestination is populated only for hardlinks / symlinks, can be an absolute or relative
	LinkDestination string
	// Size of the file in bytes
	Size     int64
	UserID   int
	GroupID  int
	Type     Type
	IsDir    bool
	Mode     os.FileMode
	MIMEType string
}

func NewMetadata(header tar.Header, content io.Reader) Metadata {
	return Metadata{
		Path:            path.Clean(DirSeparator + header.Name),
		Type:            TypeFromTarType(header.Typeflag),
		LinkDestination: header.Linkname,
		Size:            header.FileInfo().Size(),
		Mode:            header.FileInfo().Mode(),
		UserID:          header.Uid,
		GroupID:         header.Gid,
		IsDir:           header.FileInfo().IsDir(),
		MIMEType:        MIMEType(content),
	}
}

// NewMetadataFromSquashFSFile populates Metadata for the entry at path, with details from f.
func NewMetadataFromSquashFSFile(path string, f *squashfs.File) (Metadata, error) {
	fi, err := f.Stat()
	if err != nil {
		return Metadata{}, err
	}

	var ty Type
	switch {
	case fi.IsDir():
		ty = TypeDirectory
	case f.IsRegular():
		ty = TypeRegular
	case f.IsSymlink():
		ty = TypeSymLink
	default:
		switch fi.Mode() & os.ModeType {
		case os.ModeNamedPipe:
			ty = TypeFIFO
		case os.ModeSocket:
			ty = TypeSocket
		case os.ModeDevice:
			ty = TypeBlockDevice
		case os.ModeCharDevice:
			ty = TypeCharacterDevice
		case os.ModeIrregular:
			ty = TypeIrregular
		}
		// note: cannot determine hardlink from squashfs.File (but case us not possible)
	}

	md := Metadata{
		Path:            filepath.Clean(filepath.Join("/", path)),
		LinkDestination: f.SymlinkPath(),
		Size:            fi.Size(),
		IsDir:           f.IsDir(),
		Mode:            fi.Mode(),
		Type:            ty,
	}

	if f.IsRegular() {
		md.MIMEType = MIMEType(f)
	}

	return md, nil
}

func NewMetadataFromPath(path string, info os.FileInfo) Metadata {
	var mimeType string
	uid, gid := getXid(info)

	ty := TypeFromMode(info.Mode())

	if ty == TypeRegular {
		f, err := os.Open(path)
		if err != nil {
			// TODO: it may be that the file is inaccessible, however, this is not an error or a warning. In the future we need to track these as known-unknowns
			f = nil
		} else {
			defer func() {
				if err := f.Close(); err != nil {
					log.Warnf("unable to close file while obtaining metadata: %s", path)
				}
			}()
		}

		mimeType = MIMEType(f)
	}

	return Metadata{
		Path: path,
		Mode: info.Mode(),
		Type: ty,
		// unsupported across platforms
		UserID:   uid,
		GroupID:  gid,
		Size:     info.Size(),
		MIMEType: mimeType,
		IsDir:    info.IsDir(),
	}
}
