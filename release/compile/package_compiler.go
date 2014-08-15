package compile

import (
	"os"
	"path"

	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageCompiler interface {
	Compile(*bmrel.Package) error
}

type packageCompiler struct {
	runner      boshsys.CmdRunner
	packagesDir string
	fileSystem  boshsys.FileSystem
	compressor  boshcmd.Compressor
	blobstore   boshblob.Blobstore
}

func NewPackageCompiler(
	runner boshsys.CmdRunner,
	packagesDir string,
	fileSystem boshsys.FileSystem,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
) PackageCompiler {
	return &packageCompiler{
		runner:      runner,
		packagesDir: packagesDir,
		fileSystem:  fileSystem,
		compressor:  compressor,
		blobstore:   blobstore,
	}
}

func (pc *packageCompiler) Compile(pkg *bmrel.Package) error {
	packageSrcDir := pkg.ExtractedPath

	installDir := path.Join(pc.packagesDir, pkg.Name)
	err := pc.fileSystem.MkdirAll(installDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating package install dir")
	}

	defer pc.fileSystem.RemoveAll(pc.packagesDir)

	if !pc.fileSystem.FileExists(path.Join(packageSrcDir, "packaging")) {
		return bosherr.New("Packaging script for package `%s' not found", pkg.Name)
	}

	cmd := boshsys.Command{
		Name: "bash",
		Args: []string{"-x", "packaging"},
		Env: map[string]string{
			"BOSH_COMPILE_TARGET":  packageSrcDir,
			"BOSH_INSTALL_TARGET":  installDir,
			"BOSH_PACKAGE_NAME":    pkg.Name,
			"BOSH_PACKAGE_VERSION": pkg.Version,
			"BOSH_PACKAGES_DIR":    pc.packagesDir,
		},
		WorkingDir: packageSrcDir,
	}
	pc.runner.RunComplexCommand(cmd)

	tarball, err := pc.compressor.CompressFilesInDir(installDir)
	defer pc.compressor.CleanUp(tarball)

	if err != nil {
		return bosherr.WrapError(err, "Compressing compiled package")
	}

	_, _, err = pc.blobstore.Create(tarball)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	return nil
}
