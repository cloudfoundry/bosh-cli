package release

import (
	"os"
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	bmerr "github.com/cloudfoundry/bosh-micro-cli/errors"
	bmrelman "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
)

type reader struct {
	tarFilePath          string
	extractedReleasePath string
	fs                   boshsys.FileSystem
	extractor            boshcmd.Compressor
}

type Reader interface {
	Read() (Release, error)
}

func NewReader(
	tarFilePath string,
	extractedReleasePath string,
	fs boshsys.FileSystem,
	extractor boshcmd.Compressor,
) *reader {
	return &reader{
		tarFilePath:          tarFilePath,
		extractedReleasePath: extractedReleasePath,
		fs:                   fs,
		extractor:            extractor,
	}
}

func (r *reader) Read() (Release, error) {
	err := r.extractor.DecompressFileToDir(r.tarFilePath, r.extractedReleasePath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Extracting release")
	}

	releaseManifestPath := path.Join(r.extractedReleasePath, "release.MF")
	releaseManifestBytes, err := r.fs.ReadFile(releaseManifestPath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Reading release manifest")
	}

	var releaseManifest bmrelman.Release
	err = candiedyaml.Unmarshal(releaseManifestBytes, &releaseManifest)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Parsing release manifest")
	}

	release, err := r.newReleaseFromManifest(releaseManifest)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Constructing release from manifest")
	}

	return release, nil
}

func (r *reader) newReleaseFromManifest(releaseManifest bmrelman.Release) (Release, error) {
	errors := []error{}
	packages, err := r.newPackagesFromManifestPackages(releaseManifest.Packages)
	if err != nil {
		errors = append(errors, bosherr.WrapError(err, "Constructing packages from manifest"))
	}

	jobs, err := r.newJobsFromManifestJobs(packages, releaseManifest.Jobs)
	if err != nil {
		errors = append(errors, bosherr.WrapError(err, "Constructing jobs from manifest"))
	}

	if len(errors) > 0 {
		return Release{}, bmerr.NewExplainableError(errors)
	}

	return Release{
		Name:    releaseManifest.Name,
		Version: releaseManifest.Version,

		CommitHash:         releaseManifest.CommitHash,
		UncommittedChanges: releaseManifest.UncommittedChanges,

		ExtractedPath: r.extractedReleasePath,

		Jobs:     jobs,
		Packages: packages,
	}, nil
}

func (r *reader) newJobsFromManifestJobs(packages []*Package, manifestJobs []bmrelman.Job) ([]Job, error) {
	jobs := []Job{}
	errors := []error{}
	for _, manifestJob := range manifestJobs {
		extractedJobPath := path.Join(r.extractedReleasePath, "extracted_jobs", manifestJob.Name)
		err := r.fs.MkdirAll(extractedJobPath, os.ModeDir|0700)
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Creating extracted job path"))
			continue
		}

		jobArchivePath := path.Join(r.extractedReleasePath, "jobs", manifestJob.Name+".tgz")
		jobReader := NewJobReader(jobArchivePath, extractedJobPath, r.extractor, r.fs)
		job, err := jobReader.Read()
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Reading job `%s' from archive", manifestJob.Name))
			continue
		}

		job.Fingerprint = manifestJob.Fingerprint
		job.Sha1 = manifestJob.Sha1
		for _, pkgName := range job.PackageNames {
			pkg, found := r.findPackageByName(packages, pkgName)
			if !found {
				return []Job{}, bosherr.New("Package not found: `%s'", pkgName)
			}
			job.Packages = append(job.Packages, pkg)
		}

		jobs = append(jobs, job)
	}

	if len(errors) > 0 {
		return []Job{}, bmerr.NewExplainableError(errors)
	}

	return jobs, nil
}

func (r *reader) findPackageByName(packages []*Package, pkgName string) (*Package, bool) {
	for _, pkg := range packages {
		if pkg.Name == pkgName {
			return pkg, true
		}
	}
	return nil, false
}

func (r *reader) newPackagesFromManifestPackages(manifestPackages []bmrelman.Package) ([]*Package, error) {
	packages := []*Package{}
	errors := []error{}
	packageRepo := NewPackageRepo()

	for _, manifestPackage := range manifestPackages {
		pkg := packageRepo.FindOrCreatePackage(manifestPackage.Name)

		extractedPackagePath := path.Join(r.extractedReleasePath, "extracted_packages", manifestPackage.Name)
		err := r.fs.MkdirAll(extractedPackagePath, os.ModeDir|0700)
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Creating extracted package path"))
			continue
		}
		packageArchivePath := path.Join(r.extractedReleasePath, "packages", manifestPackage.Name+".tgz")
		err = r.extractor.DecompressFileToDir(packageArchivePath, extractedPackagePath)
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Extracting package `%s'", manifestPackage.Name))
			continue
		}

		pkg.Fingerprint = manifestPackage.Fingerprint
		pkg.Sha1 = manifestPackage.Sha1
		pkg.ExtractedPath = extractedPackagePath

		pkg.Dependencies = []*Package{}
		for _, manifestPackageName := range manifestPackage.Dependencies {
			pkg.Dependencies = append(pkg.Dependencies, packageRepo.FindOrCreatePackage(manifestPackageName))
		}

		packages = append(packages, pkg)
	}

	if len(errors) > 0 {
		return []*Package{}, bmerr.NewExplainableError(errors)
	}

	return packages, nil
}
