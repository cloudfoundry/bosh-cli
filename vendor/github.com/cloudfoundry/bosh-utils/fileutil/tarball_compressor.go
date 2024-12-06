package fileutil

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/klauspost/pgzip"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const forwardSlash string = "/"

type tarballCompressor struct {
	fs boshsys.FileSystem
}

func NewTarballCompressor(
	fs boshsys.FileSystem,
) Compressor {
	return tarballCompressor{fs: fs}
}

func (c tarballCompressor) CompressFilesInDir(dir string) (string, error) {
	return c.CompressSpecificFilesInDir(dir, []string{"."})
}

func (c tarballCompressor) CompressSpecificFilesInDir(dir string, files []string) (string, error) {
	tarball, err := c.fs.TempFile("bosh-platform-disk-TarballCompressor-CompressSpecificFilesInDir")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating temporary file for tarball")
	}

	defer tarball.Close()

	zw := pgzip.NewWriter(tarball)
	tw := tar.NewWriter(zw)

	for _, file := range files {
		err = c.fs.Walk(filepath.Join(dir, file), func(f string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Base(f) == ".DS_Store" {
				return nil
			}

			link := ""
			if fi.Mode()&fs.ModeSymlink != 0 {
				link, err = os.Readlink(f)
				if err != nil {
					return bosherr.WrapError(err, "Reading symlink target")
				}
			}

			header, err := tar.FileInfoHeader(fi, link)
			if err != nil {
				return bosherr.WrapError(err, "Reading tar header")
			}

			relPath, err := filepath.Rel(dir, f)
			if err != nil {
				return bosherr.WrapError(err, "Resovling relative tar path")
			}

			header.Name = relPath
			if runtime.GOOS == "windows" {
				header.Name = strings.ReplaceAll(relPath, "\\", forwardSlash)
			}

			if fi.IsDir() && header.Name[len(header.Name)-1:] != forwardSlash {
				header.Name = header.Name + forwardSlash
			}

			if len(header.Name) < 2 || header.Name[0:2] != fmt.Sprintf(".%s", forwardSlash) {
				header.Name = fmt.Sprintf(".%s%s", forwardSlash, header.Name)
			}

			if err := tw.WriteHeader(header); err != nil {
				return bosherr.WrapError(err, "Writing tar header")
			}

			if fi.Mode().IsRegular() {
				data, err := c.fs.OpenFile(f, os.O_RDONLY, 0)
				if err != nil {
					return bosherr.WrapError(err, "Reading tar source file")
				}
				defer data.Close()

				if _, err := io.Copy(tw, data); err != nil {
					return bosherr.WrapError(err, "Copying data into tar")
				}
			}
			return nil
		})
	}

	if err != nil {
		return "", bosherr.WrapError(err, "Creating tgz")
	}

	if err = tw.Close(); err != nil {
		return "", bosherr.WrapError(err, "Closing tar writer")
	}

	if err = zw.Close(); err != nil {
		return "", bosherr.WrapError(err, "Closing gzip writer")
	}

	return tarball.Name(), nil
}

func (c tarballCompressor) DecompressFileToDir(tarballPath string, dir string, options CompressorOptions) error {
	if _, err := c.fs.Stat(dir); os.IsNotExist(err) {
		return bosherr.WrapError(err, "Determine target dir")
	}

	tarball, err := c.fs.OpenFile(tarballPath, os.O_RDONLY, 0)
	if err != nil {
		return bosherr.WrapError(err, "Opening tarball")
	}
	defer tarball.Close()

	zr, err := pgzip.NewReader(tarball)
	if err != nil {
		return bosherr.WrapError(err, "Creating gzip reader")
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return bosherr.WrapError(err, "Loading next file header")
		}

		if options.PathInArchive != "" && !strings.HasPrefix(
			filepath.Clean(header.Name), filepath.Clean(options.PathInArchive)) {
			continue
		}

		fullName := filepath.Join(dir, filepath.FromSlash(header.Name))

		if options.StripComponents > 0 {
			components := strings.Split(header.Name, forwardSlash)
			if len(components) <= options.StripComponents {
				continue
			}

			fullName = filepath.Join(append([]string{dir}, components[options.StripComponents:]...)...)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := c.fs.MkdirAll(fullName, fs.FileMode(header.Mode)); err != nil {
				return bosherr.WrapError(err, "Decompressing directory")
			}

		case tar.TypeReg:
			directoryPath := filepath.Dir(fullName)
			if err := c.fs.MkdirAll(directoryPath, fs.FileMode(0755)); err != nil {
				return bosherr.WrapError(err, "Creating directory for decompressed file")
			}

			outFile, err := c.fs.OpenFile(fullName, os.O_CREATE|os.O_WRONLY, fs.FileMode(header.Mode))
			if err != nil {
				return bosherr.WrapError(err, "Creating decompressed file")
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tr); err != nil {
				return bosherr.WrapError(err, "Decompressing file contents")
			}

		case tar.TypeLink:
			if err := c.fs.Symlink(header.Linkname, fullName); err != nil {
				return bosherr.WrapError(err, "Decompressing link")
			}

		case tar.TypeSymlink:
			if err := c.fs.Symlink(header.Linkname, fullName); err != nil {
				return bosherr.WrapError(err, "Decompressing symlink")
			}

		default:
			return fmt.Errorf("unknown type: %v in %s for tar: %s",
				header.Typeflag, header.Name, tarballPath)
		}

		if options.SameOwner {
			if err := c.fs.Chown(fullName, fmt.Sprintf("%s:%s", header.Uname, header.Gname)); err != nil {
				return bosherr.WrapError(err, "Updating ownership")
			}
		}
	}
	return nil
}

func (c tarballCompressor) CleanUp(tarballPath string) error {
	return c.fs.RemoveAll(tarballPath)
}
