package pkg

import (
	gopath "path"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	. "github.com/cloudfoundry/bosh-init/release/pkg/manifest"
	. "github.com/cloudfoundry/bosh-init/release/resource"
)

type DirReaderImpl struct {
	archiveFactory ArchiveFunc

	srcDirPath   string
	blobsDirPath string

	fs boshsys.FileSystem
}

func NewDirReaderImpl(
	archiveFactory ArchiveFunc,
	srcDirPath string,
	blobsDirPath string,
	fs boshsys.FileSystem,
) DirReaderImpl {
	return DirReaderImpl{
		archiveFactory: archiveFactory,
		srcDirPath:     srcDirPath,
		blobsDirPath:   blobsDirPath,
		fs:             fs,
	}
}

func (r DirReaderImpl) Read(path string) (*Package, error) {
	manifest, files, prepFiles, err := r.collectFiles(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Collecting package files")
	}

	// Note that files do not include package's spec file,
	// but rather specify dependencies as additional chunks for the fingerprint.
	archive := r.archiveFactory(files, prepFiles, manifest.Dependencies)

	fp, err := archive.Fingerprint()
	if err != nil {
		return nil, err
	}

	resource := NewResource(manifest.Name, fp, archive)

	return NewPackage(resource, manifest.Dependencies), nil
}

func (r DirReaderImpl) collectFiles(path string) (Manifest, []File, []File, error) {
	var files, prepFiles []File

	specPath := gopath.Join(path, "spec")

	manifest, err := NewManifestFromPath(specPath, r.fs)
	if err != nil {
		return Manifest{}, nil, nil, err
	}

	packagingPath := gopath.Join(path, "packaging")

	if r.fs.FileExists(packagingPath) {
		file := NewFile(packagingPath, path)
		file.ExcludeMode = true
		files = append(files, file)
	} else {
		return manifest, nil, nil, bosherr.Errorf(
			"Expected to find '%s' for package '%s'", packagingPath, manifest.Name)
	}

	prePackagingPath := gopath.Join(path, "pre_packaging")

	if r.fs.FileExists(prePackagingPath) {
		file := NewFile(prePackagingPath, path)
		file.ExcludeMode = true
		files = append(files, file)
		prepFiles = append(prepFiles, file)
	}

	filesByRelPath := map[string]File{}

	for _, glob := range manifest.Files {
		srcDirMatches, err := r.fs.Glob(gopath.Join(r.srcDirPath, glob))
		if err != nil {
			return manifest, nil, nil, bosherr.WrapErrorf(err, "Listing package files in src")
		}

		for _, path := range srcDirMatches {
			file := NewFile(path, r.srcDirPath)
			if _, found := filesByRelPath[file.RelativePath]; !found {
				filesByRelPath[file.RelativePath] = file
			}
		}

		blobsDirMatches, err := r.fs.Glob(gopath.Join(r.blobsDirPath, glob))
		if err != nil {
			return manifest, nil, nil, bosherr.WrapErrorf(err, "Listing package files in blobs")
		}

		for _, path := range blobsDirMatches {
			file := NewFile(path, r.blobsDirPath)
			if _, found := filesByRelPath[file.RelativePath]; !found {
				filesByRelPath[file.RelativePath] = file
			}
		}
	}

	var excludedFiles []File

	for _, glob := range manifest.ExcludedFiles {
		srcDirMatches, err := r.fs.Glob(gopath.Join(r.srcDirPath, glob))
		if err != nil {
			return manifest, nil, nil, bosherr.WrapErrorf(err, "Listing package excluded files in src")
		}

		for _, path := range srcDirMatches {
			file := NewFile(path, r.srcDirPath)
			excludedFiles = append(excludedFiles, file)
		}

		blobsDirMatches, err := r.fs.Glob(gopath.Join(r.blobsDirPath, glob))
		if err != nil {
			return manifest, nil, nil, bosherr.WrapErrorf(err, "Listing package excluded files in blobs")
		}

		for _, path := range blobsDirMatches {
			file := NewFile(path, r.blobsDirPath)
			excludedFiles = append(excludedFiles, file)
		}
	}

	for _, excludedFile := range excludedFiles {
		delete(filesByRelPath, excludedFile.RelativePath)
	}

	for _, specialFileName := range []string{"packaging", "pre_packaging"} {
		if _, ok := filesByRelPath[specialFileName]; ok {
			errMsg := "Expected special '%s' file to not be included via 'files' key for package '%s'"
			return manifest, nil, nil, bosherr.Errorf(errMsg, specialFileName, manifest.Name)
		}
	}

	for _, file := range filesByRelPath {
		files = append(files, file)
	}

	return manifest, files, prepFiles, nil
}
