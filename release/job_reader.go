package release

import (
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type jobReader struct {
	archivePath      string
	extractedJobPath string
	extractor        boshcmd.Compressor
	fs               boshsys.FileSystem
}

func NewJobReader(
	archivePath string,
	extractedJobPath string,
	extractor boshcmd.Compressor,
	fs boshsys.FileSystem,
) *jobReader {
	return &jobReader{
		archivePath:      archivePath,
		extractedJobPath: extractedJobPath,
		extractor:        extractor,
		fs:               fs,
	}
}

func (r *jobReader) Read() (Job, error) {
	err := r.extractor.DecompressFileToDir(r.archivePath, r.extractedJobPath, boshcmd.CompressorOptions{})
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
		Properties:    jobManifest.Properties,
	}, nil
}
