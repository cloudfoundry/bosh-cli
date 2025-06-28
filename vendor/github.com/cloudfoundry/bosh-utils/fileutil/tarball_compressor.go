package fileutil

import (
	"fmt"
	"runtime"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
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

func (c tarballCompressor) CompressFilesInDir(dir string) (string, error) {
	return c.CompressSpecificFilesInDir(dir, []string{"."})
}

func (c tarballCompressor) CompressSpecificFilesInDir(dir string, files []string) (string, error) {
	tarball, err := c.fs.TempFile("bosh-platform-disk-TarballCompressor-CompressSpecificFilesInDir")
	if err != nil {
		return "", bosherr.WrapError(err, "Creating temporary file for tarball")
	}

	defer tarball.Close()

	tarballPath := tarball.Name()

	args := []string{"-czf", tarballPath, "-C", dir}
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
	args := []string{sameOwnerOption, "-xzf", resolvedTarballPath, "-C", dir}
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

func (c tarballCompressor) CleanUp(tarballPath string) error {
	return c.fs.RemoveAll(tarballPath)
}
