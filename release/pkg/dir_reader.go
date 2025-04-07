package pkg

import (
	"errors"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cloudfoundry/bosh-cli/v7/release/pkg/manifest"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

type DirReaderImpl struct {
	archiveFactory resource.ArchiveFunc

	srcDirPath   string
	blobsDirPath string

	fs boshsys.FileSystem
}

var (
	fileNotFoundError = errors.New("File Not Found") //nolint:staticcheck
)

func NewDirReaderImpl(
	archiveFactory resource.ArchiveFunc,
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
	manifestLock, err := r.collectLock(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Collecting package spec lock")
	}

	if manifestLock != nil {
		existingResource := resource.NewExistingResource(manifestLock.Name, manifestLock.Fingerprint, "")
		return NewPackage(existingResource, manifestLock.Dependencies), nil
	}

	manifestFromCollectFiles, files, prepFiles, err := r.collectFiles(path)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Collecting package files")
	}

	// Note that files do not include package's spec file,
	// but rather specify dependencies as additional chunks for the fingerprint.
	archive := r.archiveFactory(resource.ArchiveFactoryArgs{Files: files, PrepFiles: prepFiles, Chunks: manifestFromCollectFiles.Dependencies})

	fp, err := archive.Fingerprint()
	if err != nil {
		return nil, err
	}

	newResource := resource.NewResource(manifestFromCollectFiles.Name, fp, archive)

	return NewPackage(newResource, manifestFromCollectFiles.Dependencies), nil
}

func (r DirReaderImpl) collectLock(path string) (*manifest.ManifestLock, error) {
	path = filepath.Join(path, "spec.lock")

	if r.fs.FileExists(path) {
		manifestLock, err := manifest.NewManifestLockFromPath(path, r.fs)
		if err != nil {
			return nil, err
		}

		return &manifestLock, nil
	}

	return nil, nil
}

func (r DirReaderImpl) collectFiles(path string) (manifest.Manifest, []resource.File, []resource.File, error) {
	var files, prepFiles []resource.File

	specPath := filepath.Join(path, "spec")

	manifestFromPath, err := manifest.NewManifestFromPath(specPath, r.fs)
	if err != nil {
		return manifest.Manifest{}, nil, nil, err
	}

	packagingPath := filepath.Join(path, "packaging")
	files, err = r.checkAndFilterDir(packagingPath, path)
	if err != nil {
		if errors.Is(err, fileNotFoundError) {
			return manifestFromPath, nil, nil, bosherr.Errorf(
				"Expected to find '%s' for package '%s'", packagingPath, manifestFromPath.Name)
		}

		return manifestFromPath, nil, nil, bosherr.Errorf("Unexpected error occurred: %s", err)
	}

	prePackagingPath := filepath.Join(path, "pre_packaging")
	prepFiles, err = r.checkAndFilterDir(prePackagingPath, path) // can proceed if there is no pre_packaging
	if err != nil && !errors.Is(err, fileNotFoundError) {
		return manifestFromPath, nil, nil, bosherr.Errorf("Unexpected error occurred: %s", err)
	}

	files = append(files, prepFiles...)

	filesByRelPath, err := r.applyFilesPattern(manifestFromPath)
	if err != nil {
		return manifestFromPath, nil, nil, err
	}

	excludedFiles, err := r.applyExcludedFilesPattern(manifestFromPath)
	if err != nil {
		return manifestFromPath, nil, nil, err
	}

	for _, excludedFile := range excludedFiles {
		delete(filesByRelPath, excludedFile.RelativePath)
	}

	for _, specialFileName := range []string{"packaging", "pre_packaging"} {
		if _, ok := filesByRelPath[specialFileName]; ok {
			errMsg := "Expected special '%s' file to not be included via 'files' key for package '%s'"
			return manifestFromPath, nil, nil, bosherr.Errorf(errMsg, specialFileName, manifestFromPath.Name)
		}
	}

	for _, file := range filesByRelPath {
		files = append(files, file)
	}

	return manifestFromPath, files, prepFiles, nil
}

func (r DirReaderImpl) applyFilesPattern(manifest manifest.Manifest) (map[string]resource.File, error) {
	filesByRelPath := map[string]resource.File{}

	for _, glob := range manifest.Files {
		matchingFilesFound := false

		srcDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.srcDirPath, glob))
		if err != nil {
			return map[string]resource.File{}, bosherr.WrapErrorf(err, "Listing package files in src")
		}

		for _, path := range srcDirMatches {
			isPackageableFile, err := r.isPackageableFile(path)
			if err != nil {
				return map[string]resource.File{}, bosherr.WrapErrorf(err, "Checking file packageability")
			}

			if isPackageableFile {
				matchingFilesFound = true
				file := resource.NewFile(path, r.srcDirPath)
				if _, found := filesByRelPath[file.RelativePath]; !found {
					filesByRelPath[file.RelativePath] = file
				}
			}
		}

		blobsDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.blobsDirPath, glob))
		if err != nil {
			return map[string]resource.File{}, bosherr.WrapErrorf(err, "Listing package files in blobs")
		}

		for _, path := range blobsDirMatches {
			isPackageableFile, err := r.isPackageableFile(path)
			if err != nil {
				return map[string]resource.File{}, bosherr.WrapErrorf(err, "Checking file packageability")
			}

			if isPackageableFile {
				matchingFilesFound = true
				file := resource.NewFile(path, r.blobsDirPath)
				if _, found := filesByRelPath[file.RelativePath]; !found {
					filesByRelPath[file.RelativePath] = file
				}
			}
		}

		if !matchingFilesFound {
			return nil, bosherr.Errorf("Missing files for pattern '%s'", glob)
		}
	}

	return filesByRelPath, nil
}

func (r DirReaderImpl) applyExcludedFilesPattern(inputManifest manifest.Manifest) ([]resource.File, error) {
	var excludedFiles []resource.File

	for _, glob := range inputManifest.ExcludedFiles {
		srcDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.srcDirPath, glob))
		if err != nil {
			return []resource.File{}, bosherr.WrapErrorf(err, "Listing package excluded files in src")
		}

		for _, path := range srcDirMatches {
			file := resource.NewFile(path, r.srcDirPath)
			excludedFiles = append(excludedFiles, file)
		}

		blobsDirMatches, err := r.fs.RecursiveGlob(filepath.Join(r.blobsDirPath, glob))
		if err != nil {
			return []resource.File{}, bosherr.WrapErrorf(err, "Listing package excluded files in blobs")
		}

		for _, path := range blobsDirMatches {
			file := resource.NewFile(path, r.blobsDirPath)
			excludedFiles = append(excludedFiles, file)
		}
	}

	return excludedFiles, nil
}

func (r DirReaderImpl) checkAndFilterDir(packagePath, path string) ([]resource.File, error) {
	var files []resource.File

	if r.fs.FileExists(packagePath) {
		isPackageableFile, err := r.isPackageableFile(packagePath)
		if err != nil {
			return nil, err
		}

		if isPackageableFile {
			file := resource.NewFile(packagePath, path)
			file.ExcludeMode = true
			files = append(files, file)
		}
		return files, nil
	}

	return []resource.File{}, fileNotFoundError
}

func (r DirReaderImpl) isPackageableFile(path string) (bool, error) {
	info, err := r.fs.Lstat(path)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return false, nil
	}

	fullSrcDirPath, err := r.fs.ReadAndFollowLink(r.srcDirPath)
	if err != nil {
		return false, err
	}

	fileIsSymlink := info.Mode()&os.ModeSymlink != 0
	if !fileIsSymlink {
		fullPath, err := r.fs.ReadAndFollowLink(path)
		if err != nil {
			return false, err
		}

		relPath, err := filepath.Rel(r.srcDirPath, path)
		if err != nil {
			return false, err
		}

		fullRelPath, err := filepath.Rel(fullSrcDirPath, fullPath)
		if err != nil {
			return false, err
		}

		if relPath != fullRelPath {
			return false, nil
		}
	}

	return true, nil
}
