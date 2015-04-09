package pkg

import (
	"path/filepath"

	biinstallblob "github.com/cloudfoundry/bosh-init/installation/blob"
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
	blobExtractor biinstallblob.Extractor
}

func NewPackageInstaller(blobExtractor biinstallblob.Extractor) Installer {
	return &installer{
		blobExtractor: blobExtractor,
	}
}

func (pi *installer) Install(pkg CompiledPackageRef, parentDir string) error {
	return pi.blobExtractor.Extract(pkg.BlobstoreID, pkg.SHA1, filepath.Join(parentDir, pkg.Name))
}
