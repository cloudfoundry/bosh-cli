package fakes

import (
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
)

type FakeCompressor struct {
	CompressFilesInDirDir         string
	CompressFilesInDirOptions     boshcmd.CompressorOptions
	CompressFilesInDirTarballPath string
	CompressFilesInDirErr         error
	CompressFilesInDirCallBack    func()

	CompressSpecificFilesInDirDir         string
	CompressSpecificFilesInDirFiles       []string
	CompressSpecificFilesInDirOptions     boshcmd.CompressorOptions
	CompressSpecificFilesInDirTarballPath string
	CompressSpecificFilesInDirErr         error
	CompressSpecificFilesInDirCallBack    func()

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

func (fc *FakeCompressor) CompressFilesInDir(dir string, options boshcmd.CompressorOptions) (string, error) {
	fc.CompressFilesInDirDir = dir
	fc.CompressFilesInDirOptions = options
	if fc.CompressFilesInDirCallBack != nil {
		fc.CompressFilesInDirCallBack()
	}

	return fc.CompressFilesInDirTarballPath, fc.CompressFilesInDirErr
}

func (fc *FakeCompressor) CompressSpecificFilesInDir(dir string, files []string, options boshcmd.CompressorOptions) (string, error) {
	fc.CompressSpecificFilesInDirDir = dir
	fc.CompressSpecificFilesInDirFiles = files
	fc.CompressSpecificFilesInDirOptions = options
	if fc.CompressSpecificFilesInDirCallBack != nil {
		fc.CompressSpecificFilesInDirCallBack()
	}

	return fc.CompressSpecificFilesInDirTarballPath, fc.CompressSpecificFilesInDirErr
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
