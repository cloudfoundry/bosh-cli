package fakes

import (
	boshcmd "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/fileutil"
)

type FakeCompressor struct {
	CompressFilesInDirDir         string
	CompressFilesInDirTarballPath string
	CompressFilesInDirErr         error

	DecompressFileToDirTarballPaths []string
	DecompressFileToDirDirs         []string
	DecompressFileToDirOptions      []boshcmd.CompressorOptions
	DecompressFileToDirErr          error
	DecompressFileToDirCallBack     func()

	CleanUpTarballPath string
	CleanUpErr         error
}

func NewFakeCompressor() *FakeCompressor {
	return &FakeCompressor{}
}

func (fc *FakeCompressor) CompressFilesInDir(dir string) (string, error) {
	fc.CompressFilesInDirDir = dir
	return fc.CompressFilesInDirTarballPath, fc.CompressFilesInDirErr
}

func (fc *FakeCompressor) DecompressFileToDir(tarballPath string, dir string, options boshcmd.CompressorOptions) (err error) {
	fc.DecompressFileToDirTarballPaths = append(fc.DecompressFileToDirTarballPaths, tarballPath)
	fc.DecompressFileToDirDirs = append(fc.DecompressFileToDirDirs, dir)
	fc.DecompressFileToDirOptions = append(fc.DecompressFileToDirOptions, options)

	if fc.DecompressFileToDirCallBack != nil {
		fc.DecompressFileToDirCallBack()
	}

	return fc.DecompressFileToDirErr
}

func (fc *FakeCompressor) CleanUp(tarballPath string) error {
	fc.CleanUpTarballPath = tarballPath
	return fc.CleanUpErr
}
