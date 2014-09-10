package install

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type JobInstaller interface {
	Install(bmrel.Job, string) error
}

type jobInstaller struct {
	fs                boshsys.FileSystem
	packageInstaller  PackageInstaller
	templateExtractor BlobExtractor
	templateRepo      bmtemcomp.TemplatesRepo
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
			return bosherr.WrapError(err, "Installing package `%s'", pkg.Name)
		}
	}

	template, found, err := i.templateRepo.Find(job)
	if err != nil {
		return bosherr.WrapError(err, "Finding template for job `%s'", job.Name)
	}
	if !found {
		return bosherr.New("Could not find template for job `%s'", job.Name)
	}

	err = i.templateExtractor.Extract(template.BlobID, template.BlobSha1, jobDir)
	if err != nil {
		return bosherr.WrapError(err, "Extracting blob with ID `%s'", template.BlobID)
	}
	return nil
}

func NewJobInstaller(
	fs boshsys.FileSystem,
	packageInstaller PackageInstaller,
	blobExtractor BlobExtractor,
	templateRepo bmtemcomp.TemplatesRepo,
) JobInstaller {
	return jobInstaller{
		fs:                fs,
		packageInstaller:  packageInstaller,
		templateExtractor: blobExtractor,
		templateRepo:      templateRepo,
	}
}
