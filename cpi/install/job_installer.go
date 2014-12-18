package install

import (
	"os"
	"path"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtemcomp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler"
)

type InstalledJob struct {
	Name string
	Path string
}

type JobInstaller interface {
	Install(bmrel.Job) (InstalledJob, error)
}

type jobInstaller struct {
	fs                boshsys.FileSystem
	packageInstaller  PackageInstaller
	templateExtractor BlobExtractor
	templateRepo      bmtemcomp.TemplatesRepo
	jobsPath          string
	packagesPath      string
	eventLogger       bmeventlog.EventLogger
	timeService       boshtime.Service
}

func (i jobInstaller) Install(job bmrel.Job) (installedJob InstalledJob, err error) {
	eventLoggerStage := i.eventLogger.NewStage("installing CPI jobs")
	eventLoggerStage.Start()
	defer eventLoggerStage.Finish()

	err = eventLoggerStage.PerformStep("cpi", func() error {
		installedJob, err = i.install(job)
		return err
	})
	if err != nil {
		return InstalledJob{}, err
	}

	return installedJob, nil
}

func (i jobInstaller) install(job bmrel.Job) (InstalledJob, error) {
	jobDir := filepath.Join(i.jobsPath, job.Name)
	err := i.fs.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Creating jobs directory '%s'", jobDir)
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

func NewJobInstaller(
	fs boshsys.FileSystem,
	packageInstaller PackageInstaller,
	blobExtractor BlobExtractor,
	templateRepo bmtemcomp.TemplatesRepo,
	jobsPath,
	packagesPath string,
	eventLogger bmeventlog.EventLogger,
	timeService boshtime.Service,
) JobInstaller {
	return jobInstaller{
		fs:                fs,
		packageInstaller:  packageInstaller,
		templateExtractor: blobExtractor,
		templateRepo:      templateRepo,
		jobsPath:          jobsPath,
		packagesPath:      packagesPath,
		eventLogger:       eventLogger,
		timeService:       timeService,
	}
}
