package install

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type JobInstaller interface {
	Install(bmrel.Job, string) error
}

type jobInstaller struct {
	fs               boshsys.FileSystem
	packageInstaller PackageInstaller
}

func (i jobInstaller) Install(job bmrel.Job, path string) error {
	jobDir := filepath.Join(path, "jobs", job.Name)
	err := i.fs.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating jobs directory `%s'", jobDir)
	}

	packagesDir := filepath.Join(path, "packages")
	err = i.fs.MkdirAll(packagesDir, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating packages directory `%s'", packagesDir)
	}

	for _, pkg := range job.Packages {
		err = i.packageInstaller.Install(pkg, packagesDir)
		if err != nil {
			return bosherr.WrapError(err, "Installation failed for package `%s'", pkg.Name)
		}
	}

	// if any err, cleans things up
	return nil
}

func NewJobInstaller(
	fs boshsys.FileSystem,
	packageInstaller PackageInstaller,
) JobInstaller {
	return jobInstaller{
		fs:               fs,
		packageInstaller: packageInstaller,
	}
}
