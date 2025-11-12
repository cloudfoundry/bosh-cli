package fileutil

type CompressorOptions struct {
	SameOwner       bool
	PathInArchive   string
	StripComponents int
	NoCompression   bool
}

type Compressor interface {
	// CompressFilesInDir returns path to a compressed file
	CompressFilesInDir(dir string, options CompressorOptions) (path string, err error)

	CompressSpecificFilesInDir(dir string, files []string, options CompressorOptions) (path string, err error)

	DecompressFileToDir(path string, dir string, options CompressorOptions) (err error)

	IsNonCompressedTarball(path string) bool

	// CleanUp cleans up compressed file after it was used
	CleanUp(path string) error
}
