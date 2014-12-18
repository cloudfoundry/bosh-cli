package compile

import (
	"fmt"
	"os"
	"path"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmcpiinstall "github.com/cloudfoundry/bosh-micro-cli/cpi/install"
	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/cpi/packages"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageCompiler interface {
	Compile(*bmrel.Package) error
}

type packageCompiler struct {
	runner              boshsys.CmdRunner
	packagesDir         string
	fileSystem          boshsys.FileSystem
	compressor          boshcmd.Compressor
	blobstore           boshblob.Blobstore
	compiledPackageRepo bmpkgs.CompiledPackageRepo
	packageInstaller    bmcpiinstall.PackageInstaller
}

func NewPackageCompiler(
	runner boshsys.CmdRunner,
	packagesDir string,
	fileSystem boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	compiledPackageRepo bmpkgs.CompiledPackageRepo,
	packageInstaller bmcpiinstall.PackageInstaller,
) PackageCompiler {
	return &packageCompiler{
		runner:              runner,
		packagesDir:         packagesDir,
		fileSystem:          fileSystem,
		compressor:          compressor,
		blobstore:           blobstore,
		compiledPackageRepo: compiledPackageRepo,
		packageInstaller:    packageInstaller,
	}
}

func (pc *packageCompiler) Compile(pkg *bmrel.Package) error {
	_, found, err := pc.compiledPackageRepo.Find(*pkg)
	if err != nil {
		return bosherr.WrapError(err, fmt.Sprintf("Attempting to find compiled package '%s'", pkg.Name))
	}
	if found {
		return nil
	}

	for _, pkg := range pkg.Dependencies {
		err = pc.packageInstaller.Install(pkg, pc.packagesDir)
		if err != nil {
			return bosherr.WrapErrorf(err, "Installing package '%s' into '%s'", pkg.Name, pc.packagesDir)
		}
	}

	packageSrcDir := pkg.ExtractedPath

	installDir := path.Join(pc.packagesDir, pkg.Name)
	err = pc.fileSystem.MkdirAll(installDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating package install dir")
	}

	defer pc.fileSystem.RemoveAll(pc.packagesDir)

	if !pc.fileSystem.FileExists(path.Join(packageSrcDir, "packaging")) {
		return bosherr.Errorf("Packaging script for package '%s' not found", pkg.Name)
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
		return bosherr.WrapError(err, "Compiling package")
	}

	tarball, err := pc.compressor.CompressFilesInDir(installDir)
	if err != nil {
		return bosherr.WrapError(err, "Compressing compiled package")
	}
	defer pc.compressor.CleanUp(tarball)

	blobID, blobSHA1, err := pc.blobstore.Create(tarball)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	record := bmpkgs.CompiledPackageRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	err = pc.compiledPackageRepo.Save(*pkg, record)
	if err != nil {
		return bosherr.WrapError(err, "Saving compiled package")
	}

	return nil
}
