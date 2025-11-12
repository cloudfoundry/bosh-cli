package fileutil

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

var (
	gzipMagic   = []byte{0x1f, 0x8b}
	bzip2Magic  = []byte{0x42, 0x5a, 0x68} // "BZh"
	xzMagic     = []byte{0xfd, 0x37, 0x7a, 0x58, 0x5a, 0x00}
	zstdMagic   = []byte{0x28, 0xb5, 0x2f, 0xfd}
	ustarMagic  = []byte("ustar")
	ustarOffset = 257 // Offset of the TAR magic string in the file
)

type tarballCompressor struct {
	cmdRunner boshsys.CmdRunner
	fs        boshsys.FileSystem
}

func NewTarballCompressor(
	cmdRunner boshsys.CmdRunner,
	fs boshsys.FileSystem,
) Compressor {
	return tarballCompressor{cmdRunner: cmdRunner, fs: fs}
}

func (c tarballCompressor) CompressFilesInDir(dir string, options CompressorOptions) (string, error) {
	return c.CompressSpecificFilesInDir(dir, []string{"."}, options)
}

func (c tarballCompressor) CompressSpecificFilesInDir(dir string, files []string, options CompressorOptions) (string, error) {
	tarball, err := c.fs.TempFile("bosh-platform-disk-TarballCompressor-CompressSpecificFilesInDir")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating temporary file for tarball")
	}

	defer tarball.Close()

	tarballPath := tarball.Name()

	args := []string{"-cf", tarballPath, "-C", dir}
	if !options.NoCompression {
		args = append(args, "-z")
	}
	if runtime.GOOS == "darwin" {
		args = append([]string{"--no-mac-metadata"}, args...)
	}

	for _, file := range files { //nolint:gosimple
		args = append(args, file)
	}

	_, _, _, err = c.cmdRunner.RunCommand("tar", args...)
	if err != nil {
		return "", bosherr.WrapError(err, "Shelling out to tar")
	}

	return tarballPath, nil
}

func (c tarballCompressor) DecompressFileToDir(tarballPath string, dir string, options CompressorOptions) error {
	sameOwnerOption := "--no-same-owner"
	if options.SameOwner {
		sameOwnerOption = "--same-owner"
	}

	resolvedTarballPath, err := c.fs.ReadAndFollowLink(tarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Resolving tarball path")
	}
	args := []string{sameOwnerOption, "-xf", resolvedTarballPath, "-C", dir}
	if options.StripComponents != 0 {
		args = append(args, fmt.Sprintf("--strip-components=%d", options.StripComponents))
	}

	if options.PathInArchive != "" {
		args = append(args, options.PathInArchive)
	}
	_, _, _, err = c.cmdRunner.RunCommand("tar", args...)
	if err != nil {
		return bosherr.WrapError(err, "Shelling out to tar")
	}

	return nil
}

func (c tarballCompressor) IsNonCompressedTarball(path string) bool {
	f, err := c.fs.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		// If we cannot open the file, we assume it is not compressed
		return false
	}
	defer f.Close() //nolint:errcheck

	// Read the first 512 bytes to check both compression headers and the TAR header.
	// Ignore the error from reading a partial buffer, which is fine for short files.
	buffer := make([]byte, 512)
	_, _ = f.Read(buffer) //nolint:errcheck

	// 1. Check for compression first.
	if bytes.HasPrefix(buffer, gzipMagic) ||
		bytes.HasPrefix(buffer, bzip2Magic) ||
		bytes.HasPrefix(buffer, xzMagic) ||
		bytes.HasPrefix(buffer, zstdMagic) {
		return false
	}

	// 2. If NOT compressed, check for the TAR magic string at its specific offset.
	// Ensure the buffer is long enough to contain the TAR header magic string.
	if len(buffer) > ustarOffset+len(ustarMagic) {
		magicBytes := buffer[ustarOffset : ustarOffset+len(ustarMagic)]
		if bytes.Equal(magicBytes, ustarMagic) {
			return true
		}
	}

	return false
}

func (c tarballCompressor) CleanUp(tarballPath string) error {
	return c.fs.RemoveAll(tarballPath)
}
