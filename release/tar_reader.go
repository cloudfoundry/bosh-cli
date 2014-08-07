package release

import (
	"os"
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmerr "github.com/cloudfoundry/bosh-micro-cli/errors"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"
	bmrelman "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
)

type tarReader struct {
	tarFilePath          string
	extractedReleasePath string
	fs                   boshsys.FileSystem
	extractor            bmtar.Extractor
}

func NewTarReader(
	tarFilePath string,
	extractedReleasePath string,
	fs boshsys.FileSystem,
	extractor bmtar.Extractor,
) *tarReader {
	return &tarReader{
		tarFilePath:          tarFilePath,
		extractedReleasePath: extractedReleasePath,
		fs:                   fs,
		extractor:            extractor,
	}
}

func (r *tarReader) Read() (Release, error) {
	err := r.extractor.Extract(r.tarFilePath, r.extractedReleasePath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Extracting release")
	}

	var releaseManifest bmrelman.Release
	releaseManifestPath := path.Join(r.extractedReleasePath, "release.MF")
	releaseManifestBytes, err := r.fs.ReadFile(releaseManifestPath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Reading release manifest")
	}

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

func (r *tarReader) newReleaseFromManifest(releaseManifest bmrelman.Release) (Release, error) {
	errors := []error{}
	jobs, err := r.newJobsFromManifestJobs(releaseManifest.Jobs)
	if err != nil {
		errors = append(errors, bosherr.WrapError(err, "Constructing jobs from manifest"))
	}

	packages, err := r.newPackagesFromManifestPackages(releaseManifest.Packages)
	if err != nil {
		errors = append(errors, bosherr.WrapError(err, "Constructing packages from manifest"))
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

func (r *tarReader) newJobsFromManifestJobs(manifestJobs []bmrelman.Job) ([]bmreljob.Job, error) {
	jobs := []bmreljob.Job{}
	errors := []error{}
	for _, manifestJob := range manifestJobs {
		extractedJobPath := path.Join(r.extractedReleasePath, "extracted_jobs", manifestJob.Name)
		err := r.fs.MkdirAll(extractedJobPath, os.ModeDir)
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Creating extracted job path"))
			continue
		}

		jobArchivePath := path.Join(r.extractedReleasePath, "jobs", manifestJob.Name+".tgz")
		jobReader := bmreljob.NewTarReader(jobArchivePath, extractedJobPath, r.extractor, r.fs)
		job, err := jobReader.Read()
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Reading job `%s' from archive", manifestJob.Name))
			continue
		}

		job.Version = manifestJob.Version
		job.Fingerprint = manifestJob.Fingerprint
		job.Sha1 = manifestJob.Sha1

		jobs = append(jobs, job)
	}

	if len(errors) > 0 {
		return []bmreljob.Job{}, bmerr.NewExplainableError(errors)
	}

	return jobs, nil
}

func (r *tarReader) newPackagesFromManifestPackages(manifestPackages []bmrelman.Package) ([]Package, error) {
	packages := []Package{}
	errors := []error{}
	for _, manifestPackage := range manifestPackages {
		extractedPackagePath := path.Join(r.extractedReleasePath, "extracted_packages", manifestPackage.Name)
		err := r.fs.MkdirAll(extractedPackagePath, os.ModeDir)
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Creating extracted package path"))
			continue
		}
		packageArchivePath := path.Join(r.extractedReleasePath, "packages", manifestPackage.Name+".tgz")
		err = r.extractor.Extract(packageArchivePath, extractedPackagePath)
		if err != nil {
			errors = append(errors, bosherr.WrapError(err, "Extracting package `%s'", manifestPackage.Name))
			continue
		}

		pkg := Package{
			Name:        manifestPackage.Name,
			Version:     manifestPackage.Version,
			Fingerprint: manifestPackage.Fingerprint,
			Sha1:        manifestPackage.Sha1,

			Dependencies: manifestPackage.Dependencies,

			ExtractedPath: extractedPackagePath,
		}
		packages = append(packages, pkg)
	}

	if len(errors) > 0 {
		return []Package{}, bmerr.NewExplainableError(errors)
	}

	return packages, nil
}
