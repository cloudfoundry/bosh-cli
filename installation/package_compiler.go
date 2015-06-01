package installation

import (
	biinstallpkg "github.com/cloudfoundry/bosh-init/installation/pkg"
	bireljob "github.com/cloudfoundry/bosh-init/release/job"
	bistatejob "github.com/cloudfoundry/bosh-init/state/job"
	biui "github.com/cloudfoundry/bosh-init/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"os"
)

type PackageCompiler interface {
	For([]bireljob.Job, string, biui.Stage) ([]biinstallpkg.CompiledPackageRef, error)
}

type packageCompiler struct {
	jobDependencyCompiler bistatejob.DependencyCompiler
	fs                    boshsys.FileSystem
}

func NewPackageCompiler(
	jobDependencyCompiler bistatejob.DependencyCompiler,
	fs boshsys.FileSystem,
) PackageCompiler {
	return &packageCompiler{
		jobDependencyCompiler: jobDependencyCompiler,
		fs: fs,
	}
}

func (b *packageCompiler) For(jobs []bireljob.Job, packagesPath string, stage biui.Stage) ([]biinstallpkg.CompiledPackageRef, error) {

	err := b.fs.MkdirAll(packagesPath, os.ModePerm)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating packages directory '%s'", packagesPath)
	}

	compiledPackageRefs, err := b.jobDependencyCompiler.Compile(jobs, stage)
	if err != nil {
		return nil, bosherr.WrapError(err, "Compiling job package dependencies for installation")
	}

	compiledInstallationPackageRefs := make([]biinstallpkg.CompiledPackageRef, len(compiledPackageRefs), len(compiledPackageRefs))
	for i, compiledPackageRef := range compiledPackageRefs {
		compiledInstallationPackageRefs[i] = biinstallpkg.CompiledPackageRef{
			Name:        compiledPackageRef.Name,
			Version:     compiledPackageRef.Version,
			BlobstoreID: compiledPackageRef.BlobstoreID,
			SHA1:        compiledPackageRef.SHA1,
		}
	}

	return compiledInstallationPackageRefs, nil
}
