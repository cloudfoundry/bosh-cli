package job

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bminstallblob "github.com/cloudfoundry/bosh-micro-cli/installation/blob"
	bminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type InstalledJob struct {
	Name string
	Path string
}

type Installer interface {
	Install(bmreljob.Job, bmeventlog.Stage) (InstalledJob, error)
}

type jobInstaller struct {
	fs                boshsys.FileSystem
	packageInstaller  bminstallpkg.PackageInstaller
	templateExtractor bminstallblob.Extractor
	templateRepo      bmtemcomp.TemplatesRepo
	jobsPath          string
	packagesPath      string
	timeService       boshtime.Service
}

func (i jobInstaller) Install(job bmreljob.Job, stage bmeventlog.Stage) (installedJob InstalledJob, err error) {
	stageName := fmt.Sprintf("Installing job '%s'", job.Name)
	err = stage.PerformStep(stageName, func() error {
		installedJob, err = i.install(job)
		return err
	})
	return installedJob, err
}

func (i jobInstaller) install(job bmreljob.Job) (InstalledJob, error) {
	jobDir := filepath.Join(i.jobsPath, job.Name)
	err := i.fs.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Creating job directory '%s'", jobDir)
	}

	err = i.fs.MkdirAll(i.packagesPath, os.ModePerm)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Creating packages directory '%s'", i.packagesPath)
	}

	for _, pkg := range job.Packages {
		err = i.packageInstaller.Install(pkg, i.packagesPath)
		if err != nil {
			return InstalledJob{}, bosherr.WrapErrorf(err, "Installing package '%s'", pkg.Name)
		}
	}

	template, found, err := i.templateRepo.Find(job)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Finding template for job '%s'", job.Name)
	}
	if !found {
		return InstalledJob{}, bosherr.Errorf("Could not find template for job '%s'", job.Name)
	}

	err = i.templateExtractor.Extract(template.BlobID, template.BlobSHA1, jobDir)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Extracting blob with ID '%s'", template.BlobID)
	}

	binFiles := path.Join(jobDir, "bin", "*")
	files, err := i.fs.Glob(binFiles)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Globbing %s", binFiles)
	}

	for _, file := range files {
		err = i.fs.Chmod(file, os.FileMode(0755))
		if err != nil {
			return InstalledJob{}, bosherr.WrapErrorf(err, "Making %s executable", binFiles)
		}
	}

	return InstalledJob{Name: job.Name, Path: jobDir}, nil
}

func NewInstaller(
	fs boshsys.FileSystem,
	packageInstaller bminstallpkg.PackageInstaller,
	blobExtractor bminstallblob.Extractor,
	templateRepo bmtemcomp.TemplatesRepo,
	jobsPath,
	packagesPath string,
	timeService boshtime.Service,
) Installer {
	return jobInstaller{
		fs:                fs,
		packageInstaller:  packageInstaller,
		templateExtractor: blobExtractor,
		templateRepo:      templateRepo,
		jobsPath:          jobsPath,
		packagesPath:      packagesPath,
		timeService:       timeService,
	}
}
