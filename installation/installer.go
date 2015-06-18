package installation

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/cloudfoundry/bosh-init/installation/blobextract"
	biinstallmanifest "github.com/cloudfoundry/bosh-init/installation/manifest"
	biregistry "github.com/cloudfoundry/bosh-init/registry"
	biui "github.com/cloudfoundry/bosh-init/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type InstalledJob struct {
	RenderedJobRef
	Path string
}

func NewInstalledJob(ref RenderedJobRef, path string) InstalledJob {
	return InstalledJob{RenderedJobRef: ref, Path: path}
}

type Installer interface {
	Install(biinstallmanifest.Manifest, Target, biui.Stage) (Installation, error)
	Cleanup(Installation) error
}

type installer struct {
	jobRenderer           JobRenderer
	jobResolver           JobResolver
	packageCompiler       PackageCompiler
	blobExtractor         blobextract.Extractor
	registryServerManager biregistry.ServerManager
	logger                boshlog.Logger
	logTag                string
}

func NewInstaller(
	jobRenderer JobRenderer,
	jobResolver JobResolver,
	packageCompiler PackageCompiler,
	blobExtractor blobextract.Extractor,
	registryServerManager biregistry.ServerManager,
	logger boshlog.Logger,
) Installer {
	return &installer{
		jobRenderer:           jobRenderer,
		jobResolver:           jobResolver,
		packageCompiler:       packageCompiler,
		blobExtractor:         blobExtractor,
		registryServerManager: registryServerManager,
		logger:                logger,
		logTag:                "installer",
	}
}

func (i *installer) Install(manifest biinstallmanifest.Manifest, target Target, stage biui.Stage) (Installation, error) {
	i.logger.Info(i.logTag, "Installing CPI deployment '%s'", manifest.Name)
	i.logger.Debug(i.logTag, "Installing CPI deployment '%s' with manifest: %#v", manifest.Name, manifest)

	jobs, err := i.jobResolver.From(manifest)
	if err != nil {
		return nil, bosherr.WrapError(err, "Resolving jobs from manifest")
	}

	compiledPackages, err := i.packageCompiler.For(jobs, stage)
	if err != nil {
		return nil, err
	}

	err = stage.Perform("Installing packages", func() error {
		return i.installPackages(compiledPackages, target)
	})
	if err != nil {
		return nil, err
	}

	renderedJobRefs, err := i.jobRenderer.RenderAndUploadFrom(manifest, jobs, stage)
	renderedCPIJob := renderedJobRefs[0]
	installedJob, err := i.installJob(renderedCPIJob, target, stage)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Installing job '%s' for CPI release", renderedCPIJob.Name)
	}

	return NewInstallation(
		target,
		installedJob,
		manifest,
		i.registryServerManager,
	), nil
}

func (i *installer) Cleanup(installation Installation) error {
	job := installation.Job()
	return i.blobExtractor.Cleanup(job.BlobstoreID, job.Path)
}

func (i *installer) installPackages(compiledPackages []CompiledPackageRef, target Target) error {
	for _, pkg := range compiledPackages {
		err := i.blobExtractor.Extract(pkg.BlobstoreID, pkg.SHA1, filepath.Join(target.PackagesPath(), pkg.Name))
		if err != nil {
			return bosherr.WrapErrorf(err, "Installing package '%s'", pkg.Name)
		}
	}
	return nil
}

func (i *installer) installJob(renderedJobRef RenderedJobRef, target Target, stage biui.Stage) (installedJob InstalledJob, err error) {
	err = stage.Perform(fmt.Sprintf("Installing job '%s'", renderedJobRef.Name), func() error {
		var stageErr error
		jobDir := filepath.Join(target.JobsPath(), renderedJobRef.Name)

		stageErr = i.blobExtractor.Extract(renderedJobRef.BlobstoreID, renderedJobRef.SHA1, jobDir)
		if stageErr != nil {
			return bosherr.WrapErrorf(stageErr, "Extracting blob with ID '%s'", renderedJobRef.BlobstoreID)
		}

		stageErr = i.blobExtractor.ChmodExecutables(path.Join(jobDir, "bin", "*"))
		if stageErr != nil {
			return bosherr.WrapErrorf(stageErr, "Chmoding binaries for '%s'", jobDir)
		}

		installedJob = NewInstalledJob(renderedJobRef, jobDir)
		return nil
	})
	return installedJob, err
}
