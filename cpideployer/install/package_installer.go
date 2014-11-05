package install

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/cpideployer/packages"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageInstaller interface {
	Install(pkg *bmrel.Package, targetDir string) error
}

type packageInstaller struct {
	repo          bmpkgs.CompiledPackageRepo
	blobExtractor BlobExtractor
}

func NewPackageInstaller(repo bmpkgs.CompiledPackageRepo, blobExtractor BlobExtractor) PackageInstaller {
	return &packageInstaller{
		repo:          repo,
		blobExtractor: blobExtractor,
	}
}

func (pi *packageInstaller) Install(pkg *bmrel.Package, parentDir string) error {
	pkgRecord, found, err := pi.repo.Find(*pkg)
	if err != nil {
		return bosherr.WrapError(err, "Finding compiled package record: %#v", pkg)
	}
	if !found {
		return bosherr.New("Compiled package record not found: %#v", pkg)
	}

	return pi.blobExtractor.Extract(pkgRecord.BlobID, pkgRecord.BlobSHA1, filepath.Join(parentDir, pkg.Name))
}
