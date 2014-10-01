package install

import (
	"os"
	"path"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	boshtime "github.com/cloudfoundry/bosh-agent/time"

	bmlog "github.com/cloudfoundry/bosh-micro-cli/logging"
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
	eventLogger       bmlog.EventLogger
	timeService       boshtime.Service
}

func (i jobInstaller) Install(job bmrel.Job) (InstalledJob, error) {
	event := bmlog.Event{
		Time:  i.timeService.Now(),
		Stage: "installing CPI jobs",
		Total: 1,
		State: bmlog.Started,
		Index: 1,
		Task:  "cpi",
	}
	logErr := i.eventLogger.AddEvent(event)
	if logErr != nil {
		return InstalledJob{}, bosherr.WrapError(logErr, "Logging event: %#v", event)
	}

	installedJob, err := i.install(job)
	if err != nil {
		event = bmlog.Event{
			Time:    i.timeService.Now(),
			Stage:   "installing CPI jobs",
			Total:   1,
			State:   bmlog.Failed,
			Index:   1,
			Task:    "cpi",
			Message: err.Error(),
		}
		logErr = i.eventLogger.AddEvent(event)
		if logErr != nil {
			return InstalledJob{}, bosherr.WrapError(logErr, "Logging event: %#v", event)
		}
		return InstalledJob{}, err
	}

	event = bmlog.Event{
		Time:  i.timeService.Now(),
		Stage: "installing CPI jobs",
		Total: 1,
		State: bmlog.Finished,
		Index: 1,
		Task:  "cpi",
	}
	logErr = i.eventLogger.AddEvent(event)
	if logErr != nil {
		return InstalledJob{}, bosherr.WrapError(logErr, "Logging event: %#v", event)
	}
	return installedJob, nil
}

func (i jobInstaller) install(job bmrel.Job) (InstalledJob, error) {
	jobDir := filepath.Join(i.jobsPath, job.Name)
	err := i.fs.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return InstalledJob{}, bosherr.WrapError(err, "Creating jobs directory `%s'", jobDir)
	}

	err = i.fs.MkdirAll(i.packagesPath, os.ModePerm)
	if err != nil {
		return InstalledJob{}, bosherr.WrapError(err, "Creating packages directory `%s'", i.packagesPath)
	}

	for _, pkg := range job.Packages {
		err = i.packageInstaller.Install(pkg, i.packagesPath)
		if err != nil {
			return InstalledJob{}, bosherr.WrapError(err, "Installing package `%s'", pkg.Name)
		}
	}

	template, found, err := i.templateRepo.Find(job)
	if err != nil {
		return InstalledJob{}, bosherr.WrapError(err, "Finding template for job `%s'", job.Name)
	}
	if !found {
		return InstalledJob{}, bosherr.New("Could not find template for job `%s'", job.Name)
	}

	err = i.templateExtractor.Extract(template.BlobID, template.BlobSHA1, jobDir)
	if err != nil {
		return InstalledJob{}, bosherr.WrapError(err, "Extracting blob with ID `%s'", template.BlobID)
	}

	binFiles := path.Join(jobDir, "bin", "*")
	files, err := i.fs.Glob(binFiles)
	if err != nil {
		return InstalledJob{}, bosherr.WrapError(err, "Globbing %s", binFiles)
	}

	for _, file := range files {
		err = i.fs.Chmod(file, os.FileMode(0755))
		if err != nil {
			return InstalledJob{}, bosherr.WrapError(err, "Making %s executable", binFiles)
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
	eventLogger bmlog.EventLogger,
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
