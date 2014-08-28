package install

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	"os"

	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/packages"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

const logTag = "packageInstaller"

type PackageInstaller interface {
	Install(pkg *bmrel.Package, targetDir string) error
}

type packageInstaller struct {
	repo      bmpkgs.CompiledPackageRepo
	blobstore boshblob.Blobstore
	extractor boshcmd.Compressor
	fs        boshsys.FileSystem
	logger    boshlog.Logger
}

func NewPackageInstaller(
	repo bmpkgs.CompiledPackageRepo,
	blobstore boshblob.Blobstore,
	extractor boshcmd.Compressor,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) PackageInstaller {
	return &packageInstaller{
		repo:      repo,
		blobstore: blobstore,
		extractor: extractor,
		fs:        fs,
		logger:    logger,
	}
}

func (pi *packageInstaller) Install(pkg *bmrel.Package, targetDir string) error {
	pgkRecord, found, err := pi.repo.Find(*pkg)
	if err != nil {
		return bosherr.WrapError(err, "Finding compiled package record: %#v", pkg)
	}
	if !found {
		return bosherr.New("Compiled package record not found: %#v", pkg)
	}

	filePath, err := pi.blobstore.Get(pgkRecord.BlobID, pgkRecord.Fingerprint)
	if err != nil {
		return bosherr.WrapError(err, "Getting compiled package from blobstore: %#v", pgkRecord)
	}
	defer pi.cleanUpBlob(filePath)

	existed := pi.fs.FileExists(targetDir)
	if !existed {
		err = pi.fs.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return bosherr.WrapError(err, "Creating target dir: %s", targetDir)
		}
	}

	err = pi.extractor.DecompressFileToDir(filePath, targetDir)
	if err != nil {
		if !existed {
			pi.cleanUpFile(targetDir)
		}
		return bosherr.WrapError(err, "Extracting compiled package: %#v", pgkRecord)
	}
	return nil
}

func (pi *packageInstaller) cleanUpBlob(filePath string) {
	err := pi.blobstore.CleanUp(filePath)
	if err != nil {
		pi.logger.Error(
			logTag,
			bosherr.WrapError(err, "Removing compiled package tarball: %s", filePath).Error(),
		)
	}
}

func (pi *packageInstaller) cleanUpFile(filePath string) {
	err := pi.fs.RemoveAll(filePath)
	if err != nil {
		pi.logger.Error(
			logTag,
			bosherr.WrapError(err, "Removing: %s", filePath).Error(),
		)
	}
}
