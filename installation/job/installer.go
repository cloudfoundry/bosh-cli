package job

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bminstallblob "github.com/cloudfoundry/bosh-init/installation/blob"
	bmtemcomp "github.com/cloudfoundry/bosh-init/templatescompiler"
	bmui "github.com/cloudfoundry/bosh-init/ui"
)

type RenderedJobRef struct {
	Name        string
	Version     string
	BlobstoreID string
	SHA1        string
}

type InstalledJob struct {
	Name string
	Path string
}

type Installer interface {
	Install(RenderedJobRef, bmui.Stage) (InstalledJob, error)
}

func NewInstaller(
	fs boshsys.FileSystem,
	templateExtractor bminstallblob.Extractor,
	templateRepo bmtemcomp.TemplatesRepo,
	jobsPath string,
) Installer {
	return jobInstaller{
		fs:                fs,
		templateExtractor: templateExtractor,
		templateRepo:      templateRepo,
		jobsPath:          jobsPath,
	}
}

type jobInstaller struct {
	fs                boshsys.FileSystem
	templateExtractor bminstallblob.Extractor
	templateRepo      bmtemcomp.TemplatesRepo
	jobsPath          string
}

func (i jobInstaller) Install(renderedJobRef RenderedJobRef, stage bmui.Stage) (installedJob InstalledJob, err error) {
	stageName := fmt.Sprintf("Installing job '%s'", renderedJobRef.Name)
	err = stage.Perform(stageName, func() error {
		installedJob, err = i.install(renderedJobRef)
		return err
	})
	return installedJob, err
}

func (i jobInstaller) install(renderedJobRef RenderedJobRef) (InstalledJob, error) {
	jobDir := filepath.Join(i.jobsPath, renderedJobRef.Name)
	err := i.fs.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Creating job directory '%s'", jobDir)
	}

	err = i.templateExtractor.Extract(renderedJobRef.BlobstoreID, renderedJobRef.SHA1, jobDir)
	if err != nil {
		return InstalledJob{}, bosherr.WrapErrorf(err, "Extracting blob with ID '%s'", renderedJobRef.BlobstoreID)
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

	return InstalledJob{Name: renderedJobRef.Name, Path: jobDir}, nil
}
