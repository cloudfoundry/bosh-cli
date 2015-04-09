package pkg

import (
	"path/filepath"

	bminstallblob "github.com/cloudfoundry/bosh-init/installation/blob"
)

type CompiledPackageRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}

type Installer interface {
	Install(pkg CompiledPackageRef, targetDir string) error
}

type installer struct {
	blobExtractor bminstallblob.Extractor
}

func NewPackageInstaller(blobExtractor bminstallblob.Extractor) Installer {
	return &installer{
		blobExtractor: blobExtractor,
	}
}

func (pi *installer) Install(pkg CompiledPackageRef, parentDir string) error {
	return pi.blobExtractor.Extract(pkg.BlobstoreID, pkg.SHA1, filepath.Join(parentDir, pkg.Name))
}
