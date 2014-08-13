package jobs

import (
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
)

type tarReader struct {
	archivePath      string
	extractedJobPath string
	extractor        bmtar.Extractor
	fs               boshsys.FileSystem
}

func NewTarReader(
	archivePath string,
	extractedJobPath string,
	extractor bmtar.Extractor,
	fs boshsys.FileSystem,
) *tarReader {
	return &tarReader{
		archivePath:      archivePath,
		extractedJobPath: extractedJobPath,
		extractor:        extractor,
		fs:               fs,
	}
}

func (r *tarReader) Read() (Job, error) {
	err := r.extractor.Extract(r.archivePath, r.extractedJobPath)
	if err != nil {
		return Job{}, bosherr.WrapError(err,
			"Extracting job archive `%s'",
			r.archivePath)
	}

	jobManifestBytes, err := r.fs.ReadFile(path.Join(r.extractedJobPath, "job.MF"))
	if err != nil {
		return Job{}, bosherr.WrapError(err, "Reading job manifest")
	}

	var jobManifest Manifest
	err = candiedyaml.Unmarshal(jobManifestBytes, &jobManifest)
	if err != nil {
		return Job{}, bosherr.WrapError(err, "Parsing job manifest")
	}

	return Job{
		Name:          jobManifest.Name,
		Templates:     jobManifest.Templates,
		PackageNames:  jobManifest.Packages,
		ExtractedPath: r.extractedJobPath,
	}, nil
}
