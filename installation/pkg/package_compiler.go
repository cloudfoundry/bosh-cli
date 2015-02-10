package pkg

import (
	"os"
	"path"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"
)

type PackageCompiler interface {
	Compile(*bmrelpkg.Package) (CompiledPackageRecord, error)
}

type packageCompiler struct {
	runner              boshsys.CmdRunner
	packagesDir         string
	fileSystem          boshsys.FileSystem
	compressor          boshcmd.Compressor
	blobstore           boshblob.Blobstore
	compiledPackageRepo CompiledPackageRepo
	packageInstaller    PackageInstaller
	logger              boshlog.Logger
	logTag              string
}

func NewPackageCompiler(
	runner boshsys.CmdRunner,
	packagesDir string,
	fileSystem boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	compiledPackageRepo CompiledPackageRepo,
	packageInstaller PackageInstaller,
	logger boshlog.Logger,
) PackageCompiler {
	return &packageCompiler{
		runner:              runner,
		packagesDir:         packagesDir,
		fileSystem:          fileSystem,
		compressor:          compressor,
		blobstore:           blobstore,
		compiledPackageRepo: compiledPackageRepo,
		packageInstaller:    packageInstaller,
		logger:              logger,
		logTag:              "packageCompiler",
	}
}

func (pc *packageCompiler) Compile(pkg *bmrelpkg.Package) (record CompiledPackageRecord, err error) {
	pc.logger.Debug(pc.logTag, "Checking for compiled package '%s/%s'", pkg.Name, pkg.Fingerprint)
	record, found, err := pc.compiledPackageRepo.Find(*pkg)
	if err != nil {
		return record, bosherr.WrapErrorf(err, "Attempting to find compiled package '%s'", pkg.Name)
	}
	if found {
		return record, nil
	}

	pc.logger.Debug(pc.logTag, "Installing dependencies of package '%s/%s'", pkg.Name, pkg.Fingerprint)
	err = pc.installPackages(pkg.Dependencies)
	if err != nil {
		return record, bosherr.WrapErrorf(err, "Installing dependencies of package '%s'", pkg.Name)
	}
	defer pc.fileSystem.RemoveAll(pc.packagesDir)

	pc.logger.Debug(pc.logTag, "Compiling package '%s/%s'", pkg.Name, pkg.Fingerprint)
	installDir := path.Join(pc.packagesDir, pkg.Name)
	err = pc.fileSystem.MkdirAll(installDir, os.ModePerm)
	if err != nil {
		return record, bosherr.WrapError(err, "Creating package install dir")
	}

	packageSrcDir := pkg.ExtractedPath
	if !pc.fileSystem.FileExists(path.Join(packageSrcDir, "packaging")) {
		return record, bosherr.Errorf("Packaging script for package '%s' not found", pkg.Name)
	}

	cmd := boshsys.Command{
		Name: "bash",
		Args: []string{"-x", "packaging"},
		Env: map[string]string{
			"BOSH_COMPILE_TARGET": packageSrcDir,
			"BOSH_INSTALL_TARGET": installDir,
			"BOSH_PACKAGE_NAME":   pkg.Name,
			"BOSH_PACKAGES_DIR":   pc.packagesDir,
			"PATH":                "/usr/local/bin:/usr/bin:/bin",
		},
		UseIsolatedEnv: true,
		WorkingDir:     packageSrcDir,
	}

	_, _, _, err = pc.runner.RunComplexCommand(cmd)
	if err != nil {
		return record, bosherr.WrapError(err, "Compiling package")
	}

	tarball, err := pc.compressor.CompressFilesInDir(installDir)
	if err != nil {
		return record, bosherr.WrapError(err, "Compressing compiled package")
	}
	defer pc.compressor.CleanUp(tarball)

	blobID, blobSHA1, err := pc.blobstore.Create(tarball)
	if err != nil {
		return record, bosherr.WrapError(err, "Creating blob")
	}

	record = CompiledPackageRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	err = pc.compiledPackageRepo.Save(*pkg, record)
	if err != nil {
		return record, bosherr.WrapError(err, "Saving compiled package")
	}

	return record, nil
}

func (pc *packageCompiler) installPackages(packages []*bmrelpkg.Package) error {
	for _, pkg := range packages {
		pc.logger.Debug(pc.logTag, "Checking for compiled package '%s/%s'", pkg.Name, pkg.Fingerprint)
		record, found, err := pc.compiledPackageRepo.Find(*pkg)
		if err != nil {
			return bosherr.WrapErrorf(err, "Attempting to find compiled package '%s'", pkg.Name)
		}
		if !found {
			return bosherr.Errorf("Finding compiled package '%s'", pkg.Name)
		}

		pc.logger.Debug(pc.logTag, "Installing package '%s/%s'", pkg.Name, pkg.Fingerprint)
		compiledPackageRef := CompiledPackageRef{
			Name:        pkg.Name,
			Version:     pkg.Fingerprint,
			BlobstoreID: record.BlobID,
			SHA1:        record.BlobSHA1,
		}

		err = pc.packageInstaller.Install(compiledPackageRef, pc.packagesDir)
		if err != nil {
			return bosherr.WrapErrorf(err, "Installing package '%s' into '%s'", pkg.Name, pc.packagesDir)
		}
	}

	return nil
}
