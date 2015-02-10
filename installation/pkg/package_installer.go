package pkg

import (
	"path/filepath"

	bminstallblob "github.com/cloudfoundry/bosh-micro-cli/installation/blob"
)

type CompiledPackageRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}

//TODO: rename PackageInstaller to Installer to avoid stuttering

type PackageInstaller interface {
	Install(pkg CompiledPackageRef, targetDir string) error
}

type packageInstaller struct {
	blobExtractor bminstallblob.Extractor
}

func NewPackageInstaller(blobExtractor bminstallblob.Extractor) PackageInstaller {
	return &packageInstaller{
		blobExtractor: blobExtractor,
	}
}

func (pi *packageInstaller) Install(pkg CompiledPackageRef, parentDir string) error {
	return pi.blobExtractor.Extract(pkg.BlobstoreID, pkg.SHA1, filepath.Join(parentDir, pkg.Name))
}
