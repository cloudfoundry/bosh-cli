package file

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/anchore/stereoscope/internal/log"
	"github.com/pkg/errors"
)

const perFileReadLimit = 2 * GB

var ErrTarStopIteration = fmt.Errorf("halt iterating tar")

// tarFile is a ReadCloser of a tar file on disk.
type tarFile struct {
	io.Reader
	io.Closer
}

// TarFileEntry represents the header, contents, and list position of an entry within a tar file.
type TarFileEntry struct {
	Sequence int64
	Header   tar.Header
	Reader   io.Reader
}

// TarFileVisitor is a visitor function meant to be used in conjunction with the IterateTar.
type TarFileVisitor func(TarFileEntry) error

// ErrFileNotFound returned from ReaderFromTar if a file is not found in the given archive.
type ErrFileNotFound struct {
	Path string
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("file not found (path=%s)", e.Path)
}

// IterateTar is a function that reads across a tar and invokes a visitor function for each entry discovered. The iterator
// stops when there are no more entries to read, if there is an error in the underlying reader or visitor function,
// or if the visitor function returns a ErrTarStopIteration sentinel error.
func IterateTar(reader io.Reader, visitor TarFileVisitor) error {
	tarReader := tar.NewReader(reader)
	var sequence int64 = -1
	for {
		sequence++

		hdr, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if hdr == nil {
			continue
		}

		if err := visitor(TarFileEntry{
			Sequence: sequence,
			Header:   *hdr,
			Reader:   tarReader,
		}); err != nil {
			if errors.Is(err, ErrTarStopIteration) {
				return nil
			}
			return fmt.Errorf("failed to visit tar entry=%q : %w", hdr.Name, err)
		}
	}
	return nil
}

// ReaderFromTar returns a io.ReadCloser for the Path within a tar file.
func ReaderFromTar(reader io.ReadCloser, tarPath string) (io.ReadCloser, error) {
	var result io.ReadCloser

	visitor := func(entry TarFileEntry) error {
		if entry.Header.Name == tarPath {
			result = &tarFile{
				Reader: entry.Reader,
				Closer: reader,
			}
			return ErrTarStopIteration
		}
		return nil
	}
	if err := IterateTar(reader, visitor); err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &ErrFileNotFound{tarPath}
	}

	return result, nil
}

// MetadataFromTar returns the tar metadata from the header info.
func MetadataFromTar(reader io.ReadCloser, tarPath string) (Metadata, error) {
	var metadata *Metadata
	visitor := func(entry TarFileEntry) error {
		if entry.Header.Name == tarPath {
			var content io.Reader
			if entry.Header.Size > 0 {
				content = reader
			}
			m := NewMetadata(entry.Header, content)
			metadata = &m
			return ErrTarStopIteration
		}
		return nil
	}
	if err := IterateTar(reader, visitor); err != nil {
		return Metadata{}, err
	}
	if metadata == nil {
		return Metadata{}, &ErrFileNotFound{tarPath}
	}
	return *metadata, nil
}

// UntarToDirectory writes the contents of the given tar reader to the given destination
func UntarToDirectory(reader io.Reader, dst string) error {
	visitor := func(entry TarFileEntry) error {
		target := filepath.Join(dst, entry.Header.Name)

		switch entry.Header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(entry.Header.Mode))
			if err != nil {
				return err
			}

			// limit the reader on each file read to prevent decompression bomb attacks
			numBytes, err := io.Copy(f, io.LimitReader(entry.Reader, perFileReadLimit))
			if numBytes >= perFileReadLimit || errors.Is(err, io.EOF) {
				return fmt.Errorf("zip read limit hit (potential decompression bomb attack)")
			}
			if err != nil {
				return fmt.Errorf("unable to copy file: %w", err)
			}

			if err = f.Close(); err != nil {
				log.Errorf("failed to close file during untar of path=%q: %w", f.Name(), err)
			}
		}
		return nil
	}

	return IterateTar(reader, visitor)
}
