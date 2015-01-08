package pkg

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bminstallblob "github.com/cloudfoundry/bosh-micro-cli/installation/blob"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageInstaller interface {
	Install(pkg *bmrel.Package, targetDir string) error
}

type packageInstaller struct {
	repo          CompiledPackageRepo
	blobExtractor bminstallblob.Extractor
}

func NewPackageInstaller(repo CompiledPackageRepo, blobExtractor bminstallblob.Extractor) PackageInstaller {
	return &packageInstaller{
		repo:          repo,
		blobExtractor: blobExtractor,
	}
}

func (pi *packageInstaller) Install(pkg *bmrel.Package, parentDir string) error {
	pkgRecord, found, err := pi.repo.Find(*pkg)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding compiled package record: %#v", pkg)
	}
	if !found {
		return bosherr.Errorf("Compiled package record not found: %#v", pkg)
	}

	return pi.blobExtractor.Extract(pkgRecord.BlobID, pkgRecord.BlobSHA1, filepath.Join(parentDir, pkg.Name))
}
