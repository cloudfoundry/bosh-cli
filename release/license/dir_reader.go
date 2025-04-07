package license

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

type DirReaderImpl struct {
	archiveFactory resource.ArchiveFunc
	fs             boshsys.FileSystem
}

func NewDirReaderImpl(archiveFactory resource.ArchiveFunc, fs boshsys.FileSystem) DirReaderImpl {
	return DirReaderImpl{archiveFactory: archiveFactory, fs: fs}
}

func (r DirReaderImpl) Read(path string) (*License, error) {
	files, err := r.collectFiles(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Collecting license files")
	}

	if len(files) == 0 {
		return nil, nil
	}

	archive := r.archiveFactory(resource.ArchiveFactoryArgs{Files: files})

	fp, err := archive.Fingerprint()
	if err != nil {
		return nil, err
	}

	return NewLicense(resource.NewResource("license", fp, archive)), nil
}

func (r DirReaderImpl) collectFiles(path string) ([]resource.File, error) {
	var files []resource.File

	licenseMatches, err := r.fs.Glob(filepath.Join(path, "LICENSE*"))
	if err != nil {
		return nil, err
	}

	noticeMatches, err := r.fs.Glob(filepath.Join(path, "NOTICE*"))
	if err != nil {
		return nil, err
	}

	for _, filePath := range append(licenseMatches, noticeMatches...) {
		file := resource.NewFile(filePath, path)
		file.UseBasename = true
		file.ExcludeMode = true
		files = append(files, file)
	}

	return files, nil
}
